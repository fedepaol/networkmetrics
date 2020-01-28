package pod

import (
	"context"
	"fmt"

	"github.com/fedepaol/network-metrics/pkg/podmetrics"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_pod")

type podAction int

const (
	adding podAction = iota
	deleting
)

// Add creates a new Pod Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	log.Info("Adding pod controller")
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePod{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pod-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Pod
	// TODO Watch only those for this particular node
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	log.Info("Watching pods")

	return nil
}

// blank assignment to verify that ReconcilePod implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcilePod{}

// ReconcilePod reconciles a Pod object
type ReconcilePod struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile is the pod's controller reconcile loop
func (r *ReconcilePod) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Pod")

	// Fetch the Pod instance
	pod := &corev1.Pod{}
	err := r.client.Get(context.TODO(), request.NamespacedName, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			publishMetricsForPod(pod, deleting)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	reqLogger.Info("Received pod", "name", pod.Name)

	var podAction = adding
	if pod.DeletionTimestamp != nil {
		reqLogger.Info("Pod deleted", "name", pod.Name)
		podAction = deleting
	}

	publishMetricsForPod(pod, podAction)
	return reconcile.Result{}, nil
}

func publishMetricsForPod(pod *corev1.Pod, action podAction) error {
	// TODO here it could happen that we receive an update of a pod. We want to track only additions /
	// deletions of pods, so a map would help

	networks, err := Networks(pod)
	if err != nil {
		return fmt.Errorf("Failed to get networks out of pod %s - %s", pod.Namespace, pod.Name)
	}

	for _, n := range networks {
		if action == adding {
			podmetrics.UpdateNetAttachDefInstanceMetrics(n.PodName, n.Namespace, n.Interface, n.NetworkName, podmetrics.Adding)
		} else {
			podmetrics.UpdateNetAttachDefInstanceMetrics(n.PodName, n.Namespace, n.Interface, n.NetworkName, podmetrics.Deleting)
		}
	}
	return nil
}

package customautoscaler

import (
	"context"
	chintanbv1alpha1 "custom-autoscaler-operator/pkg/apis/chintanb/v1alpha1"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	//"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_customautoscaler")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new CustomAutoscaler Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCustomAutoscaler{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("customautoscaler-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CustomAutoscaler
	err = c.Watch(&source.Kind{Type: &chintanbv1alpha1.CustomAutoscaler{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner CustomAutoscaler
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &chintanbv1alpha1.CustomAutoscaler{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCustomAutoscaler{}

// ReconcileCustomAutoscaler reconciles a CustomAutoscaler object
type ReconcileCustomAutoscaler struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a CustomAutoscaler object and makes changes based on the state read
// and what is in the CustomAutoscaler.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCustomAutoscaler) Default_Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CustomAutoscaler")

	// Fetch the CustomAutoscaler instance
	instance := &chintanbv1alpha1.CustomAutoscaler{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		fmt.Println("Error is not nil")
		if errors.IsNotFound(err) {
			fmt.Errorf("error:", err)
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set CustomAutoscaler instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}
func int32Max(var1 int32, var2 int32 ) int32 {
	if var1 > var2 {
		return var1
	}
	return var2
}

func computeReplicas(minReplicas int32, maxReplicas int32, current_desired int32, desired_replicas int32) int32 {

	var desired int32
	desired = current_desired
	if desired == 0 {
		desired = int32Max(desired_replicas, 1)
	}
	if desired < minReplicas && minReplicas > 0 {
		desired = minReplicas
	} else if desired > maxReplicas && maxReplicas > 0 {
		desired = maxReplicas
	}
	return desired % 10 + 1
}

func scaleDeployment(deploymentName string, namespace string, scale int32, client client.Client) int {
	deployment := &appsv1.Deployment{}
	namespacedName := types.NamespacedName{
		Name:      deploymentName,
		Namespace: namespace,
	}
	// Fetch the Deployment instance
	err := client.Get(context.TODO(), namespacedName, deployment)
	if err != nil {
		panic(err)
	}
	fmt.Println("In scale, replicas is ", *deployment.Spec.Replicas)
	*deployment.Spec.Replicas = scale
	fmt.Printf("In scale Spec after update %+v\n", *deployment.Spec.Replicas)
	err = client.Update(context.TODO(), deployment)
	return 1
}

func (r *ReconcileCustomAutoscaler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	upscale_cooloff_minutes := 0.5
	downscale_cooloff_minutes := 0.5
	normal_requeue_seconds := 30
	error_requeue_seconds := 10
	
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CustomAutoscaler")
	
	custom_autoscaler_instance := &chintanbv1alpha1.CustomAutoscaler{}
	err := r.client.Get(context.TODO(), request.NamespacedName, custom_autoscaler_instance)
	if err != nil {
		fmt.Errorf("Error: %v", err)
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	
	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), request.NamespacedName, deployment)
	fmt.Printf("Deployment: %+v", *deployment.Spec.Replicas)
	minReplicas := custom_autoscaler_instance.Spec.MinReplicas
	maxReplicas := custom_autoscaler_instance.Spec.MaxReplicas
	desired_replicas := custom_autoscaler_instance.Spec.DesiredReplicas
	deploymentName := custom_autoscaler_instance.Spec.DeploymentName
	

	fmt.Println("Min:", string(minReplicas), " Max:", string(maxReplicas), "Name:", deploymentName)
	fmt.Println("Spec:", custom_autoscaler_instance.Spec)
	fmt.Println("Status:", custom_autoscaler_instance.Status)
	status := &custom_autoscaler_instance.Status
	current_desired := custom_autoscaler_instance.Status.Replicas

	new_desired := computeReplicas(minReplicas, maxReplicas, current_desired, desired_replicas)
	fmt.Println("new_desired", new_desired)
	status.Replicas = new_desired
	now := time.Now() 
	if new_desired < current_desired {
		fmt.Println("In downscale")
		time_diff := now.Sub(status.LastDownscale.Time).Minutes()
		if time_diff > downscale_cooloff_minutes {
			fmt.Println("Diff is between", now, status.LastDownscale.Time, time_diff)
			status.LastDownscale = chintanbv1alpha1.CustomDateTime{Time: time.Now()}
		} else {
			fmt.Println("Too quick to downscale so skipping")
			return reconcile.Result{RequeueAfter: time.Second * normal_requeue_seconds}, nil
		}
	} else if new_desired > current_desired {
		fmt.Println("In upscale")
		time_diff := now.Sub(status.LastUpscale.Time).Minutes()
		if  time_diff > upscale_cooloff_minutes {
			fmt.Println("Diff is between", now, status.LastUpscale.Time, time_diff)
			status.LastUpscale = chintanbv1alpha1.CustomDateTime{Time: time.Now()}
		} else {
			fmt.Println("Too quick to upscale so skipping")
			return reconcile.Result{RequeueAfter: time.Second * normal_requeue_seconds}, nil
		}
	}
 
	scaleDeployment(request.Name, request.Namespace, new_desired, r.client)
	if err != nil {
		reqLogger.Error(err, "failed to update the Deployment")
		return reconcile.Result{}, err
	}

	custom_autoscaler_instance.Status = *status
	fmt.Printf("Status after update %+v\n", custom_autoscaler_instance.Status)
	fmt.Printf("Spec after update %+v\n", custom_autoscaler_instance.Spec)
	err = r.client.Status().Update(context.TODO(), custom_autoscaler_instance)
	//r.client.Update(context.TODO(), custom_autoscaler_instance)
	fmt.Println("Done")
	if err != nil {
		reqLogger.Error(err, "failed to update the CustomAutoscaler")
		fmt.Errorf("Error: %v", err)
		return reconcile.Result{RequeueAfter: time.Second * error_requeue_seconds}, err
	}
	// Pod already exists - don't requeue
	return reconcile.Result{RequeueAfter: time.Second * normal_requeue_seconds}, nil
	//return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *chintanbv1alpha1.CustomAutoscaler) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

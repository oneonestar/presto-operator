/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/apis/core/validation"
	"time"

	"github.com/golang/glog"
	prestov1alpha1 "github.com/oneonestar/presto-controller/pkg/apis/prestocontroller/v1alpha1"
	clientset "github.com/oneonestar/presto-controller/pkg/client/clientset/versioned"
	prestoscheme "github.com/oneonestar/presto-controller/pkg/client/clientset/versioned/scheme"
	informers "github.com/oneonestar/presto-controller/pkg/client/informers/externalversions/prestocontroller/v1alpha1"
	listers "github.com/oneonestar/presto-controller/pkg/client/listers/prestocontroller/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	. "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "presto-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Presto is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Presto fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Presto"
	// MessageResourceSynced is the message used for an Event fired when a Presto
	// is synced successfully
	MessageResourceSynced = "Presto synced successfully"
)

// Controller is the controller implementation for Presto resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// prestoclientset is a clientset for our own API group
	prestoclientset clientset.Interface

	replicaSetLister  appslisters.ReplicaSetLister
	replicaSetsSynced cache.InformerSynced
	serviceLister     corev1listers.ServiceLister
	serviceSynced     cache.InformerSynced
	prestoLister      listers.PrestoLister
	prestoSynced      cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	prestoclientset clientset.Interface,
	replicaSetInformer appsinformers.ReplicaSetInformer,
	serviceInformer corev1informers.ServiceInformer,
	prestoInformer informers.PrestoInformer) *Controller {

	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(prestoscheme.AddToScheme(scheme.Scheme))
	glog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		prestoclientset:   prestoclientset,
		replicaSetLister:  replicaSetInformer.Lister(),
		replicaSetsSynced: replicaSetInformer.Informer().HasSynced,
		serviceLister:     serviceInformer.Lister(),
		serviceSynced:     serviceInformer.Informer().HasSynced,
		prestoLister:      prestoInformer.Lister(),
		prestoSynced:      prestoInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Prestos"),
		recorder:          recorder,
	}

	glog.Info("Setting up event handlers")
	// Set up an event handler for when Presto resources change
	prestoInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueuePresto,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueuePresto(new)
		},
	})
	// Set up an event handler for when ReplicaSet resources change. This
	// handler will lookup the owner of the given ReplicaSet, and if it is
	// owned by a Presto resource will enqueue that Presto resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	replicaSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newRS := new.(*appsv1.ReplicaSet)
			oldRS := old.(*appsv1.ReplicaSet)
			if newRS.ResourceVersion == oldRS.ResourceVersion {
				// Periodic resync will send update events for all known ReplicaSets.
				// Two different versions of the same ReplicaSet will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	glog.Info("Starting Presto controller")

	// Wait for the caches to be synced before starting workers
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.replicaSetsSynced, c.serviceSynced, c.prestoSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	glog.Info("Starting workers")
	// Launch two workers to process Presto resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	glog.Info("Started workers")
	<-stopCh
	glog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Presto resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		glog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Presto resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Presto resource with this namespace/name
	presto, err := c.prestoLister.Prestos(namespace).Get(name)
	if err != nil {
		// The Presto resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("presto '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	deploymentName := presto.Spec.ClusterName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: replicaSet name must be specified", key))
		return nil
	}

	_, err = c.createOrUpdateCoordinatorReplicaSet(presto)
	if err != nil {
		return err
	}

	workerReplicaSet, err := c.createOrUpdateWorkerReplicaSet(presto)
	if err != nil {
		return err
	}

	_, err = c.createOrUpdateCoordinatorService(presto)
	if err != nil {
		return err
	}

	// Finally, we update the status block of the Presto resource to reflect the
	// current state of the world
	err = c.updatePrestoStatus(presto, workerReplicaSet)
	if err != nil {
		return err
	}

	c.recorder.Event(presto, EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) getCoordinatorReplicaSet(presto *prestov1alpha1.Presto) (*appsv1.ReplicaSet, error) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{
		"app":        "presto-coordinator",
		"controller": presto.Name,
	}))
	replicaSets, err := c.replicaSetLister.ReplicaSets(presto.Namespace).List(selector)
	if err != nil {
		return nil, err
	}
	for _, replicaSet := range replicaSets {
		if metav1.IsControlledBy(replicaSet, presto) {
			return replicaSet, nil
		}
	}
	return nil, errors.NewNotFound(appsv1.Resource("replicasets"), "")
}

func (c *Controller) getCoordinatorService(presto *prestov1alpha1.Presto) (*Service, error) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{
		"app":        "presto-coordinator-service",
		"controller": presto.Name,
	}))
	replicaSets, err := c.serviceLister.Services(presto.Namespace).List(selector)
	if err != nil {
		return nil, err
	}
	for _, replicaSet := range replicaSets {
		if metav1.IsControlledBy(replicaSet, presto) {
			return replicaSet, nil
		}
	}
	return nil, errors.NewNotFound(appsv1.Resource("replicasets"), "")
}

func (c *Controller) getWorkerReplicaSet(presto *prestov1alpha1.Presto) (*appsv1.ReplicaSet, error) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{
		"app":        "presto-worker",
		"controller": presto.Name,
	}))
	replicaSets, err := c.replicaSetLister.ReplicaSets(presto.Namespace).List(selector)
	if err != nil {
		return nil, err
	}
	for _, replicaSet := range replicaSets {
		if metav1.IsControlledBy(replicaSet, presto) {
			return replicaSet, nil
		}
	}
	return nil, errors.NewNotFound(appsv1.Resource("replicasets"), "")
}

func (c *Controller) updatePrestoStatus(presto *prestov1alpha1.Presto, replicaSet *appsv1.ReplicaSet) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	prestoCopy := presto.DeepCopy()
	prestoCopy.Status.AvailableReplicas = replicaSet.Status.AvailableReplicas
	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Presto resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.prestoclientset.PrestocontrollerV1alpha1().Prestos(presto.Namespace).Update(prestoCopy)
	return err
}

// enqueuePresto takes a Presto resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Presto.
func (c *Controller) enqueuePresto(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Presto resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Presto resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		glog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	glog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Presto, we should not do anything more
		// with it.
		if ownerRef.Kind != "Presto" {
			return
		}

		presto, err := c.prestoLister.Prestos(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			glog.V(4).Infof("ignoring orphaned object '%s' of presto '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueuePresto(presto)
		return
	}
}

func (c *Controller) createOrUpdateCoordinatorReplicaSet(presto *prestov1alpha1.Presto) (*appsv1.ReplicaSet, error) {
	// Get the replicaSet with the name specified in Presto.spec
	replicaSet, err := c.getCoordinatorReplicaSet(presto)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		replicaSet, err = c.kubeclientset.AppsV1().ReplicaSets(presto.Namespace).Create(newReplicaSetCoordinator(presto))
		// If an error occurs during Get/Create, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return replicaSet, nil
}


func (c *Controller) createOrUpdateCoordinatorService(presto *prestov1alpha1.Presto) (*Service, error) {
	service, err := c.getCoordinatorService(presto)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		service, err = c.kubeclientset.CoreV1().Services(presto.Namespace).Create(newService(presto))
		// If an error occurs during Get/Create, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (c *Controller) createOrUpdateWorkerReplicaSet(presto *prestov1alpha1.Presto) (*appsv1.ReplicaSet, error) {
	// Get the replicaSet with the name specified in Presto.spec
	replicaSet, err := c.getWorkerReplicaSet(presto)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		replicaSet, err = c.kubeclientset.AppsV1().ReplicaSets(presto.Namespace).Create(newReplicaSetWorker(presto))
		// If an error occurs during Get/Create, we'll requeue the item so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return nil, err
		}
	}

	// If this number of the replicas on the Presto resource is specified, and the
	// number does not equal the current desired replicas on the ReplicaSet, we
	// should update the ReplicaSet resource.
	if presto.Spec.Replicas != nil && *presto.Spec.Replicas != *replicaSet.Spec.Replicas {
		klog.V(4).Infof("Presto %s replicas: %d, replicaSet replicas: %d", presto.Name, *replicaSet.Spec.Replicas, *replicaSet.Spec.Replicas)
		worker := replicaSet.DeepCopy()
		worker.Name = replicaSet.Name
		worker.Spec.Replicas = presto.Spec.Replicas
		replicaSet, err = c.kubeclientset.AppsV1().ReplicaSets(presto.Namespace).Update(worker)
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. THis could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return nil, err
	}
	return replicaSet, nil
}

// creates a new Deployment for a Presto resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Presto resource that 'owns' it.
func newReplicaSetCoordinator(presto *prestov1alpha1.Presto) *appsv1.ReplicaSet {
	labels := map[string]string{
		"app":        "presto-coordinator",
		"controller": presto.Name,
	}
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: getPodsPrefix(presto.Spec.ClusterName + "-coordinator"),
			Namespace:    presto.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(presto, schema.GroupVersionKind{
					Group:   prestov1alpha1.SchemeGroupVersion.Group,
					Version: prestov1alpha1.SchemeGroupVersion.Version,
					Kind:    "Presto",
				}),
			},
			Labels: labels,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: func() *int32 { i := int32(1); return &i }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: PodSpec{
					Containers: []Container{
						{
							Name:            "presto-coordinator",
							Image:           presto.Spec.Image,
							ImagePullPolicy: PullAlways,
							Ports:           []ContainerPort{{ContainerPort: 8080}},
							Env: []EnvVar{
								{Name: "DISCOVERY_URL", Value: "http://debug:8080"},
							},
							VolumeMounts: []VolumeMount{
								{Name: "config", MountPath: "/init/config"},
								{Name: "catalog", MountPath: "/init/config/catalog"},
							},
						},
					},
					Volumes: []Volume{
						{
							Name: "config",
							VolumeSource: VolumeSource{
								ConfigMap: &ConfigMapVolumeSource{
									LocalObjectReference: LocalObjectReference{
										Name: presto.Spec.CoordinatorConfig,
									},
								},
							},
						},
						{
							Name: "catalog",
							VolumeSource: VolumeSource{
								ConfigMap: &ConfigMapVolumeSource{
									LocalObjectReference: LocalObjectReference{
										Name: presto.Spec.CatalogConfig,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func newReplicaSetWorker(presto *prestov1alpha1.Presto) *appsv1.ReplicaSet {
	labels := map[string]string{
		"app":        "presto-worker",
		"controller": presto.Name,
	}
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: getPodsPrefix(presto.Spec.ClusterName + "-worker"),
			Namespace:    presto.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(presto, schema.GroupVersionKind{
					Group:   prestov1alpha1.SchemeGroupVersion.Group,
					Version: prestov1alpha1.SchemeGroupVersion.Version,
					Kind:    "Presto",
				}),
			},
			Labels: labels,
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: presto.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: PodSpec{
					Containers: []Container{
						{
							Name:            "presto-worker",
							Image:           presto.Spec.Image,
							ImagePullPolicy: PullAlways,
							Ports:           []ContainerPort{{ContainerPort: 8080}},
							Env: []EnvVar{
								{Name: "DISCOVERY_URL", Value: "http://debug:8080"},
							},
							VolumeMounts: []VolumeMount{
								{Name: "config", MountPath: "/init/config"},
								{Name: "catalog", MountPath: "/init/config/catalog"},
							},
						},
					},
					Volumes: []Volume{
						{
							Name: "config",
							VolumeSource: VolumeSource{
								ConfigMap: &ConfigMapVolumeSource{
									LocalObjectReference: LocalObjectReference{
										Name: presto.Spec.CoordinatorConfig,
									},
								},
							},
						},
						{
							Name: "catalog",
							VolumeSource: VolumeSource{
								ConfigMap: &ConfigMapVolumeSource{
									LocalObjectReference: LocalObjectReference{
										Name: presto.Spec.CatalogConfig,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func newService(presto *prestov1alpha1.Presto) *Service {
	labels := map[string]string{
		"app":        "presto-coordinator-service",
		"controller": presto.Name,
	}
	return &Service{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: getPodsPrefix(presto.Spec.ClusterName + "-service"),
			Namespace:    presto.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(presto, schema.GroupVersionKind{
					Group:   prestov1alpha1.SchemeGroupVersion.Group,
					Version: prestov1alpha1.SchemeGroupVersion.Version,
					Kind:    "Presto",
				}),
			},
			Labels: labels,
		},
		Spec: ServiceSpec{
			Selector: labels,
			Ports: []ServicePort{
				{
					Protocol:   ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}
}

func getPodsPrefix(controllerName string) string {
	// use the dash (if the name isn't too long) to make the pod name a bit prettier
	prefix := fmt.Sprintf("%s-", controllerName)
	if len(validation.ValidatePodName(prefix, true)) != 0 {
		prefix = controllerName
	}
	return prefix
}

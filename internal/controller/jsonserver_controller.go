/*
Copyright 2026.

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

package controller

import (
	"context"

	service "github.com/ntlaletsi70/json-server-controller/pkg"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	examplev1 "github.com/ntlaletsi70/json-server-controller/api/v1"
)

// JsonServerReconciler reconciles a JsonServer object
type JsonServerReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	JsonServerService *service.JsonServerService
}

// +kubebuilder:rbac:groups=example.com,resources=jsonservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=jsonservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=jsonservers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JsonServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *JsonServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	var js examplev1.JsonServer

	if err := r.Get(ctx, req.NamespacedName, &js); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := r.JsonServerService.Reconcile(ctx, &js)

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *JsonServerReconciler) SetupWithManager(mgr ctrl.Manager) error {

	log := mgr.GetLogger().WithName("JsonServer")

	//------------------------------------------------
	// Dependencies
	//------------------------------------------------

	statusWriter := service.NewWriter(
		mgr.GetClient(),
		log,
	)

	ensurer := service.NewEnsurer(
		mgr.GetClient(),
		mgr.GetScheme(),
		log,
	)

	//------------------------------------------------
	// Service
	//------------------------------------------------

	r.JsonServerService = service.NewJsonServerService(
		statusWriter,
		ensurer,
	)

	//------------------------------------------------
	// Controller
	//------------------------------------------------

	return ctrl.NewControllerManagedBy(mgr).
		For(&examplev1.JsonServer{}).
		Named("JsonServer").Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

/*

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

package controllers

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	batchv1alpha1 "sdewan.akraino.org/sdewan/api/v1alpha1"
	"sdewan.akraino.org/sdewan/openwrt"
)

var cnfRouteHandler = new(CNFRouteHandler)

type CNFRouteHandler struct {
}

func (m *CNFRouteHandler) GetType() string {
	return "cnfRoute"
}

func (m *CNFRouteHandler) GetName(instance runtime.Object) string {
	route := instance.(*batchv1alpha1.CNFRoute)
	return route.Name
}

func (m *CNFRouteHandler) GetFinalizer() string {
	return "rule.finalizers.sdewan.akraino.org"
}

func (m *CNFRouteHandler) GetInstance(r client.Client, ctx context.Context, req ctrl.Request) (runtime.Object, error) {
	instance := &batchv1alpha1.CNFRoute{}
	err := r.Get(ctx, req.NamespacedName, instance)
	return instance, err
}

func (m *CNFRouteHandler) Convert(instance runtime.Object, deployment appsv1.Deployment) (openwrt.IOpenWrtObject, error) {
	route := instance.(*batchv1alpha1.CNFRoute)
	openwrtroute := openwrt.SdewanRoute{
		Name:  route.Name,
		Dst:   route.Spec.Dst,
		Gw:    route.Spec.Gw,
		Dev:   route.Spec.Dev,
		Table: route.Spec.Table,
	}
	return &openwrtroute, nil
}

func (m *CNFRouteHandler) IsEqual(instance1 openwrt.IOpenWrtObject, instance2 openwrt.IOpenWrtObject) bool {
	route1 := instance1.(*openwrt.SdewanRoute)
	route2 := instance2.(*openwrt.SdewanRoute)
	return reflect.DeepEqual(*route1, *route2)
}

func (m *CNFRouteHandler) GetObject(clientInfo *openwrt.OpenwrtClientInfo, name string) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	route := openwrt.RouteClient{OpenwrtClient: openwrtClient}
	ret, err := route.GetRoute(name)
	return ret, err
}

func (m *CNFRouteHandler) CreateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	route := openwrt.RouteClient{OpenwrtClient: openwrtClient}
	obj := instance.(*openwrt.SdewanRoute)
	return route.CreateRoute(*obj)
}

func (m *CNFRouteHandler) UpdateObject(clientInfo *openwrt.OpenwrtClientInfo, instance openwrt.IOpenWrtObject) (openwrt.IOpenWrtObject, error) {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	route := openwrt.RouteClient{OpenwrtClient: openwrtClient}
	obj := instance.(*openwrt.SdewanRoute)
	return route.UpdateRoute(*obj)
}

func (m *CNFRouteHandler) DeleteObject(clientInfo *openwrt.OpenwrtClientInfo, name string) error {
	openwrtClient := openwrt.GetOpenwrtClient(*clientInfo)
	route := openwrt.RouteClient{OpenwrtClient: openwrtClient}
	return route.DeleteRoute(name)
}

func (m *CNFRouteHandler) Restart(clientInfo *openwrt.OpenwrtClientInfo) (bool, error) {
	return true, nil
}

// CNFRouteReconciler reconciles a CNFRoute object
type CNFRouteReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfroutes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.sdewan.akraino.org,resources=cnfroutes/status,verbs=get;update;patch

func (r *CNFRouteReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ProcessReconcile(r, r.Log, req, cnfRouteHandler)
}

func (r *CNFRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1alpha1.CNFRoute{}).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(GetToRequestsFunc(r, &batchv1alpha1.CNFRouteList{})),
			},
			Filter).
		Complete(r)
}

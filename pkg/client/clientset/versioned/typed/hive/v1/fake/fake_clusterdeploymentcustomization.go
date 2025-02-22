// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	hivev1 "github.com/openshift/hive/apis/hive/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterDeploymentCustomizations implements ClusterDeploymentCustomizationInterface
type FakeClusterDeploymentCustomizations struct {
	Fake *FakeHiveV1
	ns   string
}

var clusterdeploymentcustomizationsResource = schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeploymentcustomizations"}

var clusterdeploymentcustomizationsKind = schema.GroupVersionKind{Group: "hive.openshift.io", Version: "v1", Kind: "ClusterDeploymentCustomization"}

// Get takes name of the clusterDeploymentCustomization, and returns the corresponding clusterDeploymentCustomization object, and an error if there is any.
func (c *FakeClusterDeploymentCustomizations) Get(ctx context.Context, name string, options v1.GetOptions) (result *hivev1.ClusterDeploymentCustomization, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(clusterdeploymentcustomizationsResource, c.ns, name), &hivev1.ClusterDeploymentCustomization{})

	if obj == nil {
		return nil, err
	}
	return obj.(*hivev1.ClusterDeploymentCustomization), err
}

// List takes label and field selectors, and returns the list of ClusterDeploymentCustomizations that match those selectors.
func (c *FakeClusterDeploymentCustomizations) List(ctx context.Context, opts v1.ListOptions) (result *hivev1.ClusterDeploymentCustomizationList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(clusterdeploymentcustomizationsResource, clusterdeploymentcustomizationsKind, c.ns, opts), &hivev1.ClusterDeploymentCustomizationList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &hivev1.ClusterDeploymentCustomizationList{ListMeta: obj.(*hivev1.ClusterDeploymentCustomizationList).ListMeta}
	for _, item := range obj.(*hivev1.ClusterDeploymentCustomizationList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterDeploymentCustomizations.
func (c *FakeClusterDeploymentCustomizations) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(clusterdeploymentcustomizationsResource, c.ns, opts))

}

// Create takes the representation of a clusterDeploymentCustomization and creates it.  Returns the server's representation of the clusterDeploymentCustomization, and an error, if there is any.
func (c *FakeClusterDeploymentCustomizations) Create(ctx context.Context, clusterDeploymentCustomization *hivev1.ClusterDeploymentCustomization, opts v1.CreateOptions) (result *hivev1.ClusterDeploymentCustomization, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(clusterdeploymentcustomizationsResource, c.ns, clusterDeploymentCustomization), &hivev1.ClusterDeploymentCustomization{})

	if obj == nil {
		return nil, err
	}
	return obj.(*hivev1.ClusterDeploymentCustomization), err
}

// Update takes the representation of a clusterDeploymentCustomization and updates it. Returns the server's representation of the clusterDeploymentCustomization, and an error, if there is any.
func (c *FakeClusterDeploymentCustomizations) Update(ctx context.Context, clusterDeploymentCustomization *hivev1.ClusterDeploymentCustomization, opts v1.UpdateOptions) (result *hivev1.ClusterDeploymentCustomization, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(clusterdeploymentcustomizationsResource, c.ns, clusterDeploymentCustomization), &hivev1.ClusterDeploymentCustomization{})

	if obj == nil {
		return nil, err
	}
	return obj.(*hivev1.ClusterDeploymentCustomization), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeClusterDeploymentCustomizations) UpdateStatus(ctx context.Context, clusterDeploymentCustomization *hivev1.ClusterDeploymentCustomization, opts v1.UpdateOptions) (*hivev1.ClusterDeploymentCustomization, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(clusterdeploymentcustomizationsResource, "status", c.ns, clusterDeploymentCustomization), &hivev1.ClusterDeploymentCustomization{})

	if obj == nil {
		return nil, err
	}
	return obj.(*hivev1.ClusterDeploymentCustomization), err
}

// Delete takes name of the clusterDeploymentCustomization and deletes it. Returns an error if one occurs.
func (c *FakeClusterDeploymentCustomizations) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(clusterdeploymentcustomizationsResource, c.ns, name, opts), &hivev1.ClusterDeploymentCustomization{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClusterDeploymentCustomizations) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(clusterdeploymentcustomizationsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &hivev1.ClusterDeploymentCustomizationList{})
	return err
}

// Patch applies the patch and returns the patched clusterDeploymentCustomization.
func (c *FakeClusterDeploymentCustomizations) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *hivev1.ClusterDeploymentCustomization, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(clusterdeploymentcustomizationsResource, c.ns, name, pt, data, subresources...), &hivev1.ClusterDeploymentCustomization{})

	if obj == nil {
		return nil, err
	}
	return obj.(*hivev1.ClusterDeploymentCustomization), err
}

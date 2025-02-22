package manifest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	CapabilityAnnotation  = "capability.openshift.io/name"
	DefaultClusterProfile = "self-managed-high-availability"
)

// resourceId uniquely identifies a Kubernetes resource.
// It is used to identify any duplicate resources within
// a given set of manifests.
type resourceId struct {
	// Group identifies a set of API resources exposed together.
	// +optional
	Group string
	// Kind is the name of a particular object schema.
	Kind string
	// Name, sometimes used with the optional Namespace, helps uniquely identify an object.
	Name string
	// Namespace helps uniquely identify an object.
	// +optional
	Namespace string
}

// Manifest stores Kubernetes object in Raw from a file.
// It stores the id and the GroupVersionKind for
// the manifest. Raw and Obj should always be kept in sync
// such that each provides the same data but in different
// formats. To ensure Raw and Obj are always in sync, they
// should not be set directly but rather only be set by
// calling either method ManifestsFromFiles or
// ParseManifests.
type Manifest struct {
	// OriginalFilename is set to the filename this manifest was loaded from.
	// It is not guaranteed to be set or be unique, but will be set when
	// loading from disk to provide a better debug capability.
	OriginalFilename string

	id resourceId

	Raw []byte
	GVK schema.GroupVersionKind

	Obj *unstructured.Unstructured
}

func (r resourceId) equal(id resourceId) bool {
	return reflect.DeepEqual(r, id)
}

func (r resourceId) String() string {
	if len(r.Namespace) == 0 {
		return fmt.Sprintf("Group: %q Kind: %q Name: %q", r.Group, r.Kind, r.Name)
	} else {
		return fmt.Sprintf("Group: %q Kind: %q Namespace: %q Name: %q", r.Group, r.Kind, r.Namespace, r.Name)
	}
}

func (m Manifest) SameResourceID(manifest Manifest) bool {
	return m.id.equal(manifest.id)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the Manifest
// type. It unmarshals bytes of a single kubernetes object to Manifest.
func (m *Manifest) UnmarshalJSON(in []byte) error {
	if m == nil {
		return errors.New("Manifest: UnmarshalJSON on nil pointer")
	}

	// This happens when marshalling
	// <yaml>
	// ---	(this between two `---`)
	// ---
	// <yaml>
	if bytes.Equal(in, []byte("null")) {
		m.Raw = nil
		return nil
	}

	m.Raw = append(m.Raw[0:0], in...)
	udi, _, err := scheme.Codecs.UniversalDecoder().Decode(in, nil, &unstructured.Unstructured{})
	if err != nil {
		return errors.Wrapf(err, "unable to decode manifest")
	}
	ud, ok := udi.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected manifest to decode into *unstructured.Unstructured, got %T", ud)
	}
	m.GVK = ud.GroupVersionKind()
	m.Obj = ud
	m.id = resourceId{
		Group:     m.GVK.Group,
		Kind:      m.GVK.Kind,
		Namespace: m.Obj.GetNamespace(),
		Name:      m.Obj.GetName(),
	}
	return validateResourceId(m.id)
}

// Include returns an error if the manifest fails an inclusion filter and should be excluded from further
// processing by cluster version operator. Pointer arguments can be set nil to avoid excluding based on that
// filter. For example, setting profile non-nil and capabilities nil will return an error if the manifest's
// profile does not match, but will never return an error about capability issues.
func (m *Manifest) Include(excludeIdentifier *string, includeTechPreview *bool, profile *string, capabilities *configv1.ClusterVersionCapabilitiesStatus) error {

	annotations := m.Obj.GetAnnotations()
	if annotations == nil {
		return fmt.Errorf("no annotations")
	}

	if excludeIdentifier != nil {
		excludeAnnotation := fmt.Sprintf("exclude.release.openshift.io/%s", *excludeIdentifier)
		if v := annotations[excludeAnnotation]; v == "true" {
			return fmt.Errorf("%s=%s", excludeAnnotation, v)
		}
	}

	if includeTechPreview != nil {
		featureGateAnnotation := "release.openshift.io/feature-gate"
		featureGateAnnotationValue, featureGateAnnotationExists := annotations[featureGateAnnotation]
		if featureGateAnnotationValue == string(configv1.TechPreviewNoUpgrade) && !(*includeTechPreview) {
			return fmt.Errorf("tech-preview excluded, and %s=%s", featureGateAnnotation, featureGateAnnotationValue)
		}
		// never include the manifest if the feature-gate annotation is outside of allowed values (only TechPreviewNoUpgrade is currently allowed)
		if featureGateAnnotationExists && featureGateAnnotationValue != string(configv1.TechPreviewNoUpgrade) {
			return fmt.Errorf("unrecognized value %s=%s", featureGateAnnotation, featureGateAnnotationValue)
		}
	}

	if profile != nil {
		profileAnnotation := fmt.Sprintf("include.release.openshift.io/%s", *profile)
		if val, ok := annotations[profileAnnotation]; ok && val != "true" {
			return fmt.Errorf("unrecognized value %s=%s", profileAnnotation, val)
		} else if !ok {
			return fmt.Errorf("%s unset", profileAnnotation)
		}
	}

	// If there is no capabilities defined in a release then we do not need to check presence of capabilities in the manifest
	if capabilities != nil {
		return checkResourceEnablement(annotations, capabilities)
	}
	return nil
}

// checkResourceEnablement, given resource annotations and defined cluster capabilities, checks if the capability
// annotation exists. If so, each capability name is validated against the known set of capabilities. Each valid
// capability is then checked if it is disabled. If any invalid capabilities are found an error is returned listing
// all invalid capabilities. Otherwise, if any disabled capabilities are found an error is returned listing all
// disabled capabilities.
func checkResourceEnablement(annotations map[string]string, capabilities *configv1.ClusterVersionCapabilitiesStatus) error {
	caps := getManifestCapabilities(annotations)
	numCaps := len(caps)
	unknownCaps := make([]string, 0, numCaps)
	disabledCaps := make([]string, 0, numCaps)

	for _, c := range caps {
		var isKnownCap bool
		var isEnabledCap bool

		for _, knownCapability := range capabilities.KnownCapabilities {
			if c == knownCapability {
				isKnownCap = true
			}
		}
		if !isKnownCap {
			unknownCaps = append(unknownCaps, string(c))
			continue
		}
		for _, enabledCapability := range capabilities.EnabledCapabilities {
			if c == enabledCapability {
				isEnabledCap = true
			}

		}
		if !isEnabledCap {
			disabledCaps = append(disabledCaps, string(c))
		}
	}
	if len(unknownCaps) > 0 {
		return fmt.Errorf("unrecognized capability names: %s", strings.Join(unknownCaps, ", "))
	}
	if len(disabledCaps) > 0 {
		return fmt.Errorf("disabled capabilities: %s", strings.Join(disabledCaps, ", "))
	}
	return nil
}

// GetManifestCapabilities returns the manifest's capabilities.
func (m *Manifest) GetManifestCapabilities() []configv1.ClusterVersionCapability {
	annotations := m.Obj.GetAnnotations()
	if annotations == nil {
		return nil
	}
	return getManifestCapabilities(annotations)
}

func getManifestCapabilities(annotations map[string]string) []configv1.ClusterVersionCapability {
	val, ok := annotations[CapabilityAnnotation]

	// check for empty string val to avoid returning length 1 slice of the empty string
	if !ok || val == "" {
		return nil
	}
	caps := strings.Split(val, "+")
	allCaps := make([]configv1.ClusterVersionCapability, len(caps))

	for i, c := range caps {
		allCaps[i] = configv1.ClusterVersionCapability(c)
	}
	return allCaps
}

// ManifestsFromFiles reads files and returns Manifests in the same order.
// 'files' should be list of absolute paths for the manifests on disk. An
// error is returned for each manifest that defines a duplicate resource
// as compared to other manifests defined within the 'files' list.
// Duplicate resources have the same group, kind, name, and namespace.
func ManifestsFromFiles(files []string) ([]Manifest, error) {
	var manifests []Manifest
	ids := make(map[resourceId]bool)
	var errs []error
	for _, file := range files {
		file, err := os.Open(file)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "error opening %s", file.Name()))
			continue
		}
		defer file.Close()

		ms, err := ParseManifests(file)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "error parsing %s", file.Name()))
			continue
		}
		for _, m := range ms {
			m.OriginalFilename = filepath.Base(file.Name())
			err = addIfNotDuplicateResource(m, ids)
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "File %s contains", file.Name()))
			}
		}
		manifests = append(manifests, ms...)
	}

	agg := utilerrors.NewAggregate(errs)
	if agg != nil {
		return nil, fmt.Errorf("error loading manifests: %v", agg.Error())
	}

	return manifests, nil
}

// ParseManifests parses a YAML or JSON document that may contain one or more
// kubernetes resources. An error is returned if the input cannot be parsed
// or contains a duplicate resource.
func ParseManifests(r io.Reader) ([]Manifest, error) {
	theseIds := make(map[resourceId]bool)
	d := yaml.NewYAMLOrJSONDecoder(r, 1024)
	var manifests []Manifest
	for {
		m := Manifest{}
		if err := d.Decode(&m); err != nil {
			if err == io.EOF {
				return manifests, nil
			}
			return manifests, errors.Wrapf(err, "error parsing")
		}
		m.Raw = bytes.TrimSpace(m.Raw)
		if len(m.Raw) == 0 || bytes.Equal(m.Raw, []byte("null")) {
			continue
		}
		if err := addIfNotDuplicateResource(m, theseIds); err != nil {
			return manifests, err
		}
		manifests = append(manifests, m)
	}
}

// validateResourceId ensures the id contains the required fields per
// https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields.
func validateResourceId(id resourceId) error {
	if id.Kind == "" || id.Name == "" {
		return fmt.Errorf("Resource with fields %s must contain kubernetes required fields kind and name", id)
	}
	return nil
}

func addIfNotDuplicateResource(manifest Manifest, resourceIds map[resourceId]bool) error {
	if _, ok := resourceIds[manifest.id]; !ok {
		resourceIds[manifest.id] = true
		return nil
	}
	return fmt.Errorf("duplicate resource: (%s)", manifest.id)
}

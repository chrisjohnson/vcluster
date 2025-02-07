package volumesnapshotcontents

import (
	volumesnapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	"github.com/loft-sh/vcluster/pkg/util/translate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (s *syncer) translate(vVSC *volumesnapshotv1.VolumeSnapshotContent) (*volumesnapshotv1.VolumeSnapshotContent, error) {
	target, err := s.translator.Translate(vVSC)
	if err != nil {
		return nil, err
	}

	pVSC := target.(*volumesnapshotv1.VolumeSnapshotContent)
	pVSC.Spec.VolumeSnapshotRef = corev1.ObjectReference{
		Namespace: s.targetNamespace,
		Name:      translate.PhysicalName(vVSC.Spec.VolumeSnapshotRef.Name, vVSC.Spec.VolumeSnapshotRef.Namespace),
	}
	return pVSC, nil
}

func (s *syncer) translateBackwards(pVSC *volumesnapshotv1.VolumeSnapshotContent, vVS *volumesnapshotv1.VolumeSnapshot) *volumesnapshotv1.VolumeSnapshotContent {
	// build virtual VolumeSnapshotContent object
	vObj := pVSC.DeepCopy()
	vObj.ResourceVersion = ""
	vObj.UID = ""
	vObj.ManagedFields = nil

	vObj.Spec.VolumeSnapshotRef = translateVolumeSnapshotRefBackwards(&vObj.Spec.VolumeSnapshotRef, vVS)

	if vObj.Annotations == nil {
		vObj.Annotations = map[string]string{}
	}
	vObj.Annotations[HostClusterVSCAnnotation] = pVSC.Name

	return vObj
}

func (s *syncer) translateUpdateBackwards(pVSC, vVSC *volumesnapshotv1.VolumeSnapshotContent, vVS *volumesnapshotv1.VolumeSnapshot) *volumesnapshotv1.VolumeSnapshotContent {
	var updated *volumesnapshotv1.VolumeSnapshotContent

	// add a finalizer to ensure that we delete the physical VolumeSnapshotContent object when virtual is being deleted
	pCopy := pVSC.DeepCopy()
	if pCopy.Finalizers == nil {
		pCopy.Finalizers = []string{}
	}
	controllerutil.AddFinalizer(pCopy, PhysicalVSCGarbageCollectionFinalizer)

	if !equality.Semantic.DeepEqual(vVSC.Finalizers, pCopy.Finalizers) {
		updated = newIfNil(updated, vVSC)
		updated.Finalizers = pCopy.Finalizers
	}

	//TODO: consider syncing certain annotations, e.g.:
	// "snapshot.storage.kubernetes.io/volumesnapshot-being-deleted" or
	// "snapshot.storage.kubernetes.io/volumesnapshot-being-created"

	return updated
}

func (s *syncer) translateUpdate(vVSC *volumesnapshotv1.VolumeSnapshotContent, pVSC *volumesnapshotv1.VolumeSnapshotContent) *volumesnapshotv1.VolumeSnapshotContent {
	var updated *volumesnapshotv1.VolumeSnapshotContent

	if !equality.Semantic.DeepEqual(pVSC.Spec.DeletionPolicy, vVSC.Spec.DeletionPolicy) {
		updated = newIfNil(updated, pVSC)
		updated.Spec.DeletionPolicy = vVSC.Spec.DeletionPolicy
	}

	if !equality.Semantic.DeepEqual(pVSC.Spec.VolumeSnapshotClassName, vVSC.Spec.VolumeSnapshotClassName) {
		updated = newIfNil(updated, pVSC)
		updated.Spec.VolumeSnapshotClassName = vVSC.Spec.VolumeSnapshotClassName
	}

	updatedAnnotations := s.translator.TranslateAnnotations(vVSC, pVSC)
	if !equality.Semantic.DeepEqual(updatedAnnotations, pVSC.Annotations) {
		updated = newIfNil(updated, pVSC)
		updated.Annotations = updatedAnnotations
	}

	updatedLabels := s.translator.TranslateLabels(vVSC)
	if !equality.Semantic.DeepEqual(updatedLabels, pVSC.Labels) {
		updated = newIfNil(updated, pVSC)
		updated.Labels = updatedLabels
	}

	return updated
}

func translateVolumeSnapshotRefBackwards(ref *corev1.ObjectReference, vVS *volumesnapshotv1.VolumeSnapshot) corev1.ObjectReference {
	newRef := ref.DeepCopy()
	newRef.Namespace = vVS.Namespace
	newRef.Name = vVS.Name
	newRef.UID = vVS.UID
	newRef.ResourceVersion = vVS.ResourceVersion
	return *newRef
}

func newIfNil(updated *volumesnapshotv1.VolumeSnapshotContent, objBase *volumesnapshotv1.VolumeSnapshotContent) *volumesnapshotv1.VolumeSnapshotContent {
	if updated == nil {
		return objBase.DeepCopy()
	}
	return updated
}

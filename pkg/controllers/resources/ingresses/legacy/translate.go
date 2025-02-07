package legacy

import (
	"github.com/loft-sh/vcluster/pkg/util/translate"
	"github.com/pkg/errors"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func (s *syncer) translate(vIngress *networkingv1beta1.Ingress) (*networkingv1beta1.Ingress, error) {
	newObj, err := s.translator.Translate(vIngress)
	if err != nil {
		return nil, errors.Wrap(err, "error setting metadata")
	}

	newIngress := newObj.(*networkingv1beta1.Ingress)
	newIngress.Spec = *translateSpec(vIngress.Namespace, &vIngress.Spec)
	return newIngress, nil
}

func (s *syncer) translateUpdate(pObj, vObj *networkingv1beta1.Ingress) *networkingv1beta1.Ingress {
	var updated *networkingv1beta1.Ingress

	translatedSpec := *translateSpec(vObj.Namespace, &vObj.Spec)
	if !equality.Semantic.DeepEqual(translatedSpec, pObj.Spec) {
		updated = newIfNil(updated, pObj)
		updated.Spec = translatedSpec
	}

	translatedAnnotations := s.translator.TranslateAnnotations(vObj, pObj)
	if !equality.Semantic.DeepEqual(translatedAnnotations, pObj.Annotations) {
		updated = newIfNil(updated, pObj)
		updated.Annotations = translatedAnnotations
	}

	translatedLabels := s.translator.TranslateLabels(vObj)
	if !equality.Semantic.DeepEqual(translatedLabels, pObj.Labels) {
		updated = newIfNil(updated, pObj)
		updated.Labels = translatedLabels
	}

	return updated
}

func (s *syncer) translateUpdateBackwards(pObj, vObj *networkingv1beta1.Ingress) *networkingv1beta1.Ingress {
	var updated *networkingv1beta1.Ingress

	if vObj.Spec.IngressClassName == nil && pObj.Spec.IngressClassName != nil {
		updated = newIfNil(updated, vObj)
		updated.Spec.IngressClassName = pObj.Spec.IngressClassName
	}

	return updated
}

func translateSpec(namespace string, vIngressSpec *networkingv1beta1.IngressSpec) *networkingv1beta1.IngressSpec {
	retSpec := vIngressSpec.DeepCopy()
	if retSpec.Backend != nil {
		if retSpec.Backend.ServiceName != "" {
			retSpec.Backend.ServiceName = translate.PhysicalName(retSpec.Backend.ServiceName, namespace)
		}
		if retSpec.Backend.Resource != nil {
			retSpec.Backend.Resource.Name = translate.PhysicalName(retSpec.Backend.Resource.Name, namespace)
		}
	}

	for i, rule := range retSpec.Rules {
		if rule.HTTP != nil {
			for j, path := range rule.HTTP.Paths {
				if path.Backend.ServiceName != "" {
					retSpec.Rules[i].HTTP.Paths[j].Backend.ServiceName = translate.PhysicalName(retSpec.Rules[i].HTTP.Paths[j].Backend.ServiceName, namespace)
				}
				if path.Backend.Resource != nil {
					retSpec.Rules[i].HTTP.Paths[j].Backend.Resource.Name = translate.PhysicalName(retSpec.Rules[i].HTTP.Paths[j].Backend.Resource.Name, namespace)
				}
			}
		}
	}

	for i, tls := range retSpec.TLS {
		if tls.SecretName != "" {
			retSpec.TLS[i].SecretName = translate.PhysicalName(retSpec.TLS[i].SecretName, namespace)
		}
	}

	return retSpec
}

func newIfNil(updated *networkingv1beta1.Ingress, pObj *networkingv1beta1.Ingress) *networkingv1beta1.Ingress {
	if updated == nil {
		return pObj.DeepCopy()
	}
	return updated
}

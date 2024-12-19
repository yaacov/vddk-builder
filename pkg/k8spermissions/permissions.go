package k8spermissions

import (
	"context"
	"fmt"

	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// CreateClientWithToken creates a Kubernetes clientset using the provided token.
func CreateClientWithToken(apiServer, token string) (*kubernetes.Clientset, error) {
	config := &rest.Config{
		Host:        apiServer,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // Adjust if using a trusted CA
		},
	}
	return kubernetes.NewForConfig(config)
}

// CheckAccessWithToken checks if the token can perform the specified action.
func CheckAccessWithToken(clientset *kubernetes.Clientset, verb, resource string) (bool, error) {
	sar := &v1.SelfSubjectAccessReview{
		Spec: v1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &v1.ResourceAttributes{
				Verb:     verb,
				Resource: resource,
			},
		},
	}

	result, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), sar, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to create SelfSubjectAccessReview: %w", err)
	}

	return result.Status.Allowed, nil
}

PROJECT_ROOT = ../..
include ${PROJECT_ROOT}/hack/help.mk

.PHONY: test
test: ## Run integration test.
	go run main.go

.PHONY: cluster-info
cluster-info: ## Print useful info about the cluster regarding integration run
	@echo "###################"
	@echo "Pods"
	@echo "kubectl get pods -A"
	@echo "###################"
	@kubectl get pods -A || true
	@echo ""

	@echo "####################################"
	@echo "DockerRegistry CR"
	@echo "kubectl get dockerregistry -A -oyaml"
	@echo "####################################"
	@kubectl get dockerregistry -A -oyaml || true
	@echo ""

	@echo "########################################################################"
	@echo "Describe dockerregistry operator deploy"
	@echo "kubectl describe deploy -A -l app.kubernetes.io/name=serverless-operator"
	@echo "########################################################################"
	@kubectl describe deploy -A -l app.kubernetes.io/name=serverless-operator || true
	@echo ""

	@echo "###########################################################################"
	@echo "Get dockerregistry operator pod"
	@echo "kubectl get pod -A -l app.kubernetes.io/name=dockerregistry-operator -oyaml"
	@echo "###########################################################################"
	@kubectl get pod -A -l app.kubernetes.io/name=dockerregistry-operator -oyaml || true
	@echo ""

	@echo "############################################################################################################"
	@echo "Operator logs"
	@echo "kubectl logs -n kyma-system -l app.kubernetes.io/component=dockerregistry-operator.kyma-project.io --tail=-1"
	@echo "############################################################################################################"
	@kubectl logs -n kyma-system -l app.kubernetes.io/component=dockerregistry-operator.kyma-project.io --tail=-1 || true
	@echo ""

	@echo "################################################"
	@echo "Network Policies"
	@echo "kubectl get networkpolicies.networking.k8s.io -A"
	@echo "################################################"
	@kubectl get networkpolicies.networking.k8s.io -A
	@echo ""

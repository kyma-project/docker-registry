.PHONY: docker_push_simple_app
docker_push_simple_app:
	@kyma alpha registry config --output config.json
	# TODO: get multiplatform nginx; now we get nginx for amd64
	docker pull nginx@sha256:f05d105face814474acc5538160bd3f29309d9398dd895a4e71f676a4fd9a3fc
	docker tag nginx@sha256:f05d105face814474acc5538160bd3f29309d9398dd895a4e71f676a4fd9a3fc $$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.externalAccess.pushAddress})/simple-app:0.0.1
	docker --config . push $$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.externalAccess.pushAddress})/simple-app:0.0.1

.PHONY: docker_registry_login
docker_registry_login:
	@echo "Loging Docker CLI into kyma internal docker registry..."
	@docker login -u $$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.username} | base64 -d) -p $$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.password} | base64 -d) $$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.externalAccess.pushAddress})

.PHONY: deploy_simple_app
deploy_simple_app:
	@echo "Deploying simple app service..."
	@kubectl apply -f k8s-resources/simple-app
	@kubectl set image deployment simple-app-service simple-app-service=$$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.internalAccess.pullAddress})/simple-app:0.0.1
	@echo "Waiting for simple app service..."
	@kubectl wait --for condition=Available deployment/simple-app-service --timeout=180s
	@echo "Simple app service available."


.PHONY: ensure-btp-object-store-backend
ensure-btp-object-store-backend:
# create reference instance (and binding) to an existing btp object store service instance that is shared via credentials to remote service manager from different subaccount.
	@kubectl create secret generic remote-service-manager-credentials \
		--namespace kyma-system --from-env-file sm.env
	@echo "Waiting for CRD btp operator"
	@while ! kubectl get crd btpoperators.operator.kyma-project.io; \
		do echo "Waiting for CRD btp operator..."; sleep 1; done
	@kubectl wait --for condition=established crd/btpoperators.operator.kyma-project.io
	@while ! kubectl get btpoperators.operator.kyma-project.io btpoperator --namespace kyma-system; \
		do echo "Waiting for btpoperator..."; sleep 1; done
	@kubectl wait --for condition=Ready btpoperators.operator.kyma-project.io/btpoperator -n kyma-system --timeout=180s
	@kyma alpha reference-instance \
		--btp-secret-name remote-service-manager-credentials \
		--namespace kyma-system \
		--offering-name objectstore \
		--plan-selector standard \
		--reference-name object-store-reference
	@kubectl apply -n kyma-system -f k8s-resources/dependencies/object-store-binding.yaml
	@while ! kubectl get secret object-store-reference-binding --namespace kyma-system; \
		do echo "Waiting for object-store-reference-binding secret..."; sleep 5; done

.PHONY: enable_docker_registry
enable_docker_registry:
	@echo "Waiting for kyma gateway"
	@while ! kubectl get gateway kyma-gateway -n kyma-system; do echo "Waiting for kyma-gateway..."; sleep 1; done
	@echo "Enabling docker registry"
	@kubectl apply -f ${OPERATOR_MANIFEST}
	@while ! kubectl get crd dockerregistries.operator.kyma-project.io; do echo "Waiting for docker-registry CRD..."; sleep 1; done
	@kubectl wait --for condition=established crd/dockerregistries.operator.kyma-project.io
	@kubectl apply -f k8s-resources/dependencies/exposed-docker-registry-with-btp-storage.yaml -n kyma-system
	@echo "..waiting for docker registry"
	@kubectl wait --for condition=Installed dockerregistries.operator.kyma-project.io/default -n kyma-system --timeout=300s
	@echo "Docker registry installed"
	@while ! kubectl get secret dockerregistry-config-external; do echo "Waiting for dockerregistry-config-external secret..."; sleep 1; done
	@echo "Waiting for docker registry external address availibility"
	@for i in 1 2 3 4; do ./hack/registry-availibility-check.sh ${KUBECONFIG} && break || sleep 5; done

.PHONY: test
test: docker_push_simple_app deploy_simple_app cleanup-kyma

.PHONY cleanup-kyma:
cleanup-kyma:
	@echo "Deleting resources from kyma..."
	@kubectl delete -f k8s-resources/simple-app
	@kubectl delete -n kyma-system servicebindings.services.cloud.sap.com --all
	@kubectl delete -n kyma-system serviceinstances.services.cloud.sap.com --all




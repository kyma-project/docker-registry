.PHONY: headless-kubeconfig
headless-kubeconfig:
	pwd
	ls -la
	( \
	set -e -o pipefail ;\
	curl -s -o tmp/kubeconfig.yaml $$(cat ../tf/kubeconfig_url.txt) ;\
	ISSUER=$$(cat ../tf/binding_credentials.json | jq  -r '. | { clientid, url: (.url+ "/oauth2/token")  }' ) ;\
	TOKEN=$$(curl -X POST  $$(jq -r '. | .url' <<< $$ISSUER) \
	-H 'Content-Type: application/x-www-form-urlencoded' \
	-d 'grant_type=password' \
	-d 'username='"$$TF_VAR_BTP_BOT_USER" \
	-d 'password='"$$TF_VAR_BTP_BOT_PASSWORD" \
	-d 'client_id='$$(jq -r '. | .clientid' <<< $$ISSUER) \
	-d 'scope=groups, email' \
	| jq -r '.id_token' ) ;\
	echo "i got $$(jq -r '. | .clientid' <<< $$ISSUER)" ;\
	../bin/kyma alpha oidc --kubeconfig=tmp/kubeconfig.yaml --output=tmp/headless-kubeconfig.yaml --token=$$TOKEN ;\
	)

.PHONY: service-account-kubeconfig
service-account-kubeconfig:
	@echo "Generating expiring service-account based kubeconfig..."
	../bin/kyma alpha access --clusterrole ${CLUSTERROLE} --kubeconfig tmp/headless-kubeconfig.yaml --name ci-workflow-sa --time 1h --output tmp/sa-kubeconfig.yaml
	@echo "Service-account based kubeconfig generated"

.PHONY: docker_push_simple_app
docker_push_simple_app:
	# TODO: get multiplatform nginx; now we get nginx for amd64
	docker pull nginx@sha256:f05d105face814474acc5538160bd3f29309d9398dd895a4e71f676a4fd9a3fc
	docker tag nginx@sha256:f05d105face814474acc5538160bd3f29309d9398dd895a4e71f676a4fd9a3fc $$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.externalAccess.pushAddress} --kubeconfig ${KUBECONFIG})/simple-app:0.0.1
	docker push $$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.externalAccess.pushAddress} --kubeconfig ${KUBECONFIG})/simple-app:0.0.1

.PHONY: enable_docker_registry
enable_docker_registry:
	@echo "Waiting for kyma gateway"
	@while ! kubectl get gateway kyma-gateway -n kyma-system --kubeconfig ${KUBECONFIG}; do echo "Waiting for kyma-gateway..."; sleep 1; done
	@echo "Enabling docker registry"
	@kubectl apply -f ${OPERATOR_MANIFEST} --kubeconfig ${KUBECONFIG}
	@kubectl apply -f ../k8s-resources/dependencies/exposed-docker-registry-with-btp-storage.yaml -n kyma-system --kubeconfig ${KUBECONFIG}
	@echo "..waiting for docker registry"
	@kubectl wait --for condition=Installed dockerregistries.operator.kyma-project.io/default -n kyma-system --timeout=300s --kubeconfig ${KUBECONFIG}
	@echo "Docker registry installed"
	@while ! kubectl get secret dockerregistry-config-external --kubeconfig ${KUBECONFIG}; do echo "Waiting for dockerregistry-config-external secret..."; sleep 1; done
	@echo "Waiting for docker registry external address availibility"
	./registry-availibility-check.sh ${KUBECONFIG}

.PHONY: docker_registry_login
docker_registry_login:
	@echo "Loging Docker CLI into kyma internal docker registry..."
	@docker login -u $$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.username} --kubeconfig ${KUBECONFIG} | base64 -d) -p $$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.password} --kubeconfig ${KUBECONFIG} | base64 -d) $$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.externalAccess.pushAddress} --kubeconfig ${KUBECONFIG})

.PHONY: deploy_simple_app
deploy_simple_app:
	@echo "Deploying simple app service..."
	@kubectl apply -k ../k8s-resources/simple-app --kubeconfig ${KUBECONFIG}
	@kubectl set image deployment simple-app-service simple-app-service=$$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.internalAccess.pullAddress} --kubeconfig ${KUBECONFIG})/simple-app:0.0.1 --kubeconfig ${KUBECONFIG}
	@echo "Waiting for simple app service..."
	@kubectl wait --for condition=Available deployment/simple-app-service --kubeconfig ${KUBECONFIG} --timeout=180s
	@echo "Simple app service available."


.PHONY: share-sm-service-operator-access
share-sm-service-operator-access: ## update a secret used for service manager access to allow managing shared stateless service instances from a "service provider" subaccount 
# create secret --from-env-file .env in the target kyma runtime
	@kubectl create secret generic sm-provider-access --namespace kyma-system --from-env-file ../tf/provider-sm-decoded.env --kubeconfig ${KUBECONFIG}


.PHONY: create-object-store-reference
create-object-store-reference:
# create reference instance (and binding) to an existing btp object store service instance that is shared from the "provider" subaccount
	@echo "Waiting for CRD btp operator"
	@while ! kubectl get crd btpoperators.operator.kyma-project.io --kubeconfig ${KUBECONFIG}; do echo "Waiting for CRD btp operator..."; sleep 1; done
	@kubectl wait --for condition=established crd/btpoperators.operator.kyma-project.io
	@while ! kubectl get btpoperators.operator.kyma-project.io btpoperator --namespace kyma-system --kubeconfig ${KUBECONFIG}; do echo "Waiting for btpoperator..."; sleep 1; done
	@kubectl wait --for condition=Ready btpoperators.operator.kyma-project.io/btpoperator -n kyma-system --timeout=180s
	@../bin/kyma alpha reference-instance --btp-secret-name sm-provider-access --namespace kyma-system --offering-name objectstore --name-selector provider-object-store --reference-name object-store-reference --kubeconfig ${KUBECONFIG}
	@kubectl apply -n kyma-system -f ../k8s-resources/dependencies/object-store-binding.yaml --kubeconfig ${KUBECONFIG}
	@while ! kubectl get secret object-store-reference-binding --namespace kyma-system --kubeconfig ${KUBECONFIG}; do echo "Waiting for object-store-reference-binding secret..."; sleep 1; done

.PHONY cleanup-tf-force:
cleanup-tf-force:
	rm -rf ../tf/*.txt
	rm -rf ../tf/*.json
	rm -rf ../tf/*.env
	rm -rf ../tf/*.txt
	rm -rf ../tf/terraform.state
	rm -rf ../tf/terraform.state
	rm -rf ../tf/terraform.tfstate
	rm -rf ../tf/terraform.tfstate*
	rm -rf ../tf/.terraform.locl.hcl

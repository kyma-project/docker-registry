#!/bin/sh

USERNAME=$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.username} --kubeconfig ${KUBECONFIG} | base64 -d)
PASSWORD=$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.password} --kubeconfig ${KUBECONFIG} | base64 -d)
REGISTRY_URL=$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.externalAccess.pushAddress} --kubeconfig ${KUBECONFIG})


echo Testing Docker Registry availibility at: $REGISTRY_URL

# TODO: https://github.tools.sap/otters/kyma-automation-demo/issues/5
sleep 10

COUNTER=0
RESPONSE_CODE=$(curl -o /dev/null -u $USERNAME:$PASSWORD -L -w ''%{http_code}'' --connect-timeout 5 \
    --max-time 10 \
    --retry 15 \
    --retry-delay 5 \
    --retry-max-time 40 $REGISTRY_URL 2>/dev/null)
echo Response from registry: $RESPONSE_CODE
if [ "$RESPONSE_CODE" == "200" ]; then
    exit 0
fi

echo "ERROR"
exit 1

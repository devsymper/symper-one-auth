#!/bin/bash

echo "Prepare deployment directive /root/kubernetes/deployment/${SERVICE_ENV}/${SERVICE_NAME}"
ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "echo -e '$USER_PASS' | sudo -S rm -rf /root/kubernetes/deployment/${SERVICE_ENV}/${SERVICE_NAME}"
ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "echo -e '$USER_PASS' | sudo -S mkdir -p /root/kubernetes/deployment/${SERVICE_ENV}/${SERVICE_NAME}"

echo "Prepare deployment temp directive /tmp/${SERVICE_ENV}/${SERVICE_NAME}"
ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "rm -rf /tmp/${SERVICE_ENV}/${SERVICE_NAME}"
ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "mkdir -p /tmp/${SERVICE_ENV}/${SERVICE_NAME}"

echo "Upload manifests to temp directive /tmp/${SERVICE_ENV}/${SERVICE_NAME}"
scp -o StrictHostKeyChecking=no deploy/k8s/* $USER_NAME@$SSH_HOST:/tmp/${SERVICE_ENV}/${SERVICE_NAME}
ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "echo -e '$USER_PASS' | sudo -S  cp /tmp/${SERVICE_ENV}/${SERVICE_NAME}/* /root/kubernetes/deployment/${SERVICE_ENV}/${SERVICE_NAME}"

echo "Switch kubernetes namespace to ${SERVICE_ENV}"
ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "echo -e '$USER_PASS' | sudo -S  kubectl config set-context --current --namespace=${SERVICE_ENV}"

echo "Apply all manifests /root/kubernetes/deployment/${SERVICE_ENV}/${SERVICE_NAME}"
ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "echo -e '$USER_PASS' | sudo -S  kubectl apply -f /root/kubernetes/deployment/${SERVICE_ENV}/${SERVICE_NAME}"

echo "Check deployment $APP_NAME-$TARGET_ROLE status"
ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "echo -e '$USER_PASS' | sudo -S kubectl rollout status deployment/${APP_NAME}-${TARGET_ROLE}"
STATUS=$?
echo "status: $STATUS"
if [ "$STATUS" == 0 ] 
then
    echo "Deployment succeed!"
    echo "Switch traffic to new deployment verison $TARGET_ROLE"
    ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "echo -e '$USER_PASS' | sudo -S kubectl patch service $APP_NAME -p '{\"spec\": {\"selector\": {\"role\": \"$TARGET_ROLE\"}}}'"
    echo "Destroy old deployment version $APP_NAME-$CURRENT_ROLE"
    ssh -o StrictHostKeyChecking=no $USER_NAME@$SSH_HOST "echo -e '$USER_PASS' | sudo -S kubectl delete deployment $APP_NAME-$CURRENT_ROLE --ignore-not-found=true"
fi
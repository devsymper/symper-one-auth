#!/bin/bash

envDomain=""
originEnv=""

# Set originEnv and envDomain based on SERVICE_ENV
if [[ -n $SERVICE_ENV && $SERVICE_ENV != "prod" ]]; then
    envDomain="${SERVICE_ENV}-"
    originEnv=$SERVICE_ENV
fi

# Debug: Check value of envDomain
echo "ENVDomain: ${envDomain}"
echo "HostDomain: ${envDomain}${HOST_NAME}"

# Update `app_configmap.yaml`
sed -i -e "s/{APP_NAME}/$APP_NAME/g" \
       -e "s/{ENVIRONMENT}/$originEnv/g" \
       -e "s/{POSTGRES_USER}/$POSTGRES_USER/g" \
       -e "s/{POSTGRES_PASS}/$POSTGRES_PASS/g" \
       -e "s/{POSTGRES_DB}/${POSTGRES_DB}/g" \
       -e "s/{POSTGRES_HOST}/$POSTGRES_HOST/g" \
       -e "s/{CLICKHOUSE_USER}/$CLICKHOUSE_USER/g" \
       -e "s/{CLICKHOUSE_PASS}/$CLICKHOUSE_PASS/g" \
       -e "s/{CLICKHOUSE_DB}/${CLICKHOUSE_DB}/g" \
       -e "s/{CLICKHOUSE_HOST}/$CLICKHOUSE_HOST/g"\
        deploy/k8s/app_configmap.yaml

# Update `app_deployment.yaml`
sed -i -e "s|{SYMPER_IMAGE}|$IMAGE_TAG_NAME|g" \
       -e "s/{APP_NAME}/$APP_NAME/g" \
       -e "s/{SERVICE_NAME}/$SERVICE_NAME/g" \
       -e "s/{TARGET_ROLE}/$TARGET_ROLE/g"  \
        deploy/k8s/app_deployment.yaml

# Update `service_ingress.yaml`
sed -i -e "s/{APP_NAME}/$APP_NAME/g" \
       -e "s/{CURRENT_ROLE}/$CURRENT_ROLE/g" \
       -e "s/{HOST_DOMAIN}/${envDomain}${HOST_NAME}/g"  \
        deploy/k8s/service_ingress.yaml

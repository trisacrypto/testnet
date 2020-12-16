# Cluster Devops

**Kubernetes configurations for the GCP Integration Cluster**

## Installation Steps

1. Install and configure kubectl for your cluster
2. Install `helm` and configure for the cluster (v3)
3. Apply namespaces

    ```
    $ kubectl apply -f namespaces.yaml
    ```

4. Install gcp-dns-admin credentials

    Optional: create Google service account if not already done

    ```
    $ export PROJECT_ID=[project-id]
    $ gcloud iam service-accounts create dns-admin --display-name "dns-admin"
    $ gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member serviceAccount:dns-admin@$PROJECT_ID.iam.gserviceaccount.com \
        --role roles/dns.admin
    ```

    Create static credentials as a k8s secret

    ```
    $ gcloud iam service-accounts keys create creds/gcp-dns-admin.json \
        --iam-account dns-admin@$PROJECT_ID.iam.gserviceaccount.com
    $ kubectl -n cert-manager create secret generic clouddns-admin-svc-acct \
        --from-file=creds/gcp-dns-admin.json
    ```

    These credentials will enable the cert-manager dns01 solver with Google Cloud DNS.

5. Install kubed using helm

    ```
    $ helm repo add appscode https://charts.appscode.com/stable/
    $ helm repo update
    $ helm install kubed appscode/kubed \
        --version v0.12.0 \
        --namespace kube-system
    ```

    kubed is responsible for synchronizing the tls-certs secret to all namespaces with the label `app=routable`.

5. Install cert-manager using helm

    ```
    $ helm repo add jetstack https://charts.jetstack.io
    $ helm repo update
    $ helm install cert-manager jetstack/cert-manager \
        --namespace cert-manager \
        --version v1.1.0 \
        --set installCRDs=true
    ```

    Apply the cert-manger certificates

    ```
    $ kubectl apply -f cert-manager.yaml
    ```

    cert-manager performs dns01 solving for lets-encrypt cert issuance.

6. Install traefik using helm

    ```
    $ kubectl apply -f traefik/traefik-config.configmap.yaml
    $ helm repo add traefik https://helm.traefik.io/traefik
    $ helm repo update
    $ helm install traefik traefik/traefik \
        --namespace=global
        --values=traefik/chart-values.yaml
    ```

    Traefic is our external load balancer and primary ingress for the cluster.

7. Configure traefik dashboard

    Create a basic-auth secret to login

    ```
    $ htpasswd -nb admin@trisa.io [supersecretpassword] | openssl base64
    ```

    Write the results into `creds/traefik-dashboard-auth.yaml`

    ```
    apiVersion: v1
    kind: Secret
    metadata:
    name: traefik-dashboard-auth
    namespace: global
    data:
    users: |2
        YWRtaW5AdHJpc2EuaW86JGFwcjEkQ3dwRnN1RHckVXE0Y2ZiQ1hjbm9nTEwvdXBEZGhNLgoK
    ```

    Then apply the required files

    ```
    $ kubectl apply -f creds/traefik-dashboard-auth.yaml
    $ kubectl apply -f traefik/dashboard.yaml
    ```

    The traefik dashboard allows administrative views of the traefik manager.

8. Configure our current services

    ```
    $ kubectl apply -f dsweb.yaml
    $ kubectl apply -f placeholder.yaml
    $ kubectl apply -f trisads.yaml
    $ kubectl apply -f trisa.routes.yaml
    ```
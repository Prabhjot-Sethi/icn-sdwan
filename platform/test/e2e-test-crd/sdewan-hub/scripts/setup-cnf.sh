#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

base=$(pwd)

test -f $base/variables
. $base/variables
providerSubnet=${providerSubnet}
providerGateway=${providerGateway}
providerExcludeIps=${providerExcludeIps}
providerNetworkInterface=${providerNetworkInterface}
cnfWanGateway=${cnfWanGateway}

clean()
{
echo "Cleaning ..."
kubectl delete -f network-prepare.yaml
kubectl delete -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml
[-f ipsec_config.yaml ] && kubectl delete -f ipsec_config.yaml
[-f ipsec_proposal.yaml ] && kubectl delete -f ipsec_proposal.yaml
}

error_detect()
{
	echo "Error on line $1"
	#clean
}

trap "error_detect $LINENO" ERR

echo "--------------------- Setup sdewan controller ---------------------"
helm package ../../edge-scripts/helm-tmp/controllers
helm install ./controllers-0.1.0.tgz --generate-name
sleep 1m

echo "--------------------- Applying CRDs ---------------------"
cat > ipsec_proposal.yaml << EOF
---
apiVersion: batch.sdewan.akraino.org/v1alpha1
kind: IpsecProposal
metadata:
  name: ipsecproposal
  namespace: default
  labels:
    sdewanPurpose: $sdewan_cnf_name
spec:
  dh_group: modp3072
  encryption_algorithm: aes128
  hash_algorithm: sha256

EOF

kubectl apply -f ipsec_proposal.yaml

cat > ipsec_config.yaml << EOF
---
apiVersion: batch.sdewan.akraino.org/v1alpha1
kind: IpsecSite
metadata:
  name: ipsecsite
  namespace: default
  labels:
    sdewanPurpose: $sdewan_cnf_name
spec:
    name: sdewan-hub 
    remote: "%any" 
    pre_shared_key: test_key
    authentication_method: psk
    local_identifier: $hubIp
    crypto_proposal:
      - ipsecproposal
    force_crypto_proposal: "0"
    connections:
    - name: connA
      conn_type: tunnel
      mode: start
      remote_sourceip: "192.168.1.5-192.168.1.6"
      local_subnet: 192.168.1.1/24,$hubIp/32
      crypto_proposal:
        - ipsecproposal

EOF

kubectl apply -f ipsec_config.yaml

echo "--------------------- Configuration finished ---------------------"

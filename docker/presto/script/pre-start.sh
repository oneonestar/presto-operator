#!/usr/bin/env bash

rm -rf /etc/presto
cp -r /init/config /etc/presto

sed -i '$a\' /etc/presto/node.properties
echo "node.id=$(uuidgen)" >> /etc/presto/node.properties

sed -i '$a\' /etc/presto/config.properties
echo "discovery.uri=${DISCOVERY_URL}" >> /etc/presto/config.properties
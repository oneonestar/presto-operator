FROM centos:7
MAINTAINER oneonestar

ENV VERSION=302
ENV PRESTO_VERSION=${VERSION}
ENV PRESTO_HOME=/usr/local/presto
ENV BASE_URL=https://repo1.maven.org/maven2

RUN yum update -y
RUN yum install -y java-1.8.0-openjdk

RUN curl ${BASE_URL}/io/prestosql/presto-server-rpm/${PRESTO_VERSION}/presto-server-rpm-${PRESTO_VERSION}.rpm \
        -o presto-server-rpm-${PRESTO_VERSION}.rpm \
    && yum localinstall -y presto-server-rpm-${PRESTO_VERSION}.rpm \
    && rm -rf presto-server-rpm-${PRESTO_VERSION}.rpm

COPY config /init/config
COPY script /init/script

WORKDIR /init/script
ENTRYPOINT ["/bin/bash", "-c", "./start-presto.sh"]

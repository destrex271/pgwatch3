# ----------------------------------------------------------------
# 1. Build Web UI
# ----------------------------------------------------------------
FROM node:21 AS uibuilder
ADD src/webui /webui
RUN cd webui && yarn install --network-timeout 100000 && yarn build

# ----------------------------------------------------------------
# 2. Build gatherer
# ----------------------------------------------------------------
FROM golang:1.22 as builder

ARG VERSION
ARG GIT_HASH
ARG GIT_TIME

ADD src /pgwatch3
COPY --from=uibuilder /webui/build /pgwatch3/webui/build
RUN cd /pgwatch3 && CGO_ENABLED=0 go build -ldflags "-X 'main.commit=${GIT_HASH}' -X 'main.date=${GIT_TIME}' -X 'main.version=${VERSION}'"

# ----------------------------------------------------------------
# 3. Build the final image
# ----------------------------------------------------------------
FROM postgres:16-bullseye AS releaser

# Copy over the compiled gatherer
COPY --from=builder /pgwatch3/pgwatch3 /pgwatch3/
COPY src/metrics/metrics.yaml /pgwatch3/metrics/metrics.yaml

# Install dependencies
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update \
    && apt-get -qy install curl gnupg postgresql-common apt-transport-https lsb-release musl \
    && sh /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh -y \
    && echo "deb https://packagecloud.io/timescale/timescaledb/debian/ $(lsb_release -c -s) main" | tee /etc/apt/sources.list.d/timescaledb.list \
    && curl -L "https://packagecloud.io/timescale/timescaledb/gpgkey" | apt-key add - \
    && curl -L "https://www.postgresql.org/media/keys/ACCC4CF8.asc" | apt-key add - \
    && apt-get update \
    && apt-get -qy install \
        timescaledb-2-postgresql-16 postgresql-plpython3-16 postgresql-16-pg-qualstats \
        supervisor python3-psutil libfontconfig1 \
    && apt-get purge -y --auto-remove \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install Grafana
RUN arch=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) \
    && curl "https://dl.grafana.com/oss/release/grafana_10.3.3_${arch}.deb" --output grafana.deb \
    && dpkg -i grafana.deb && rm grafana.deb
COPY docker/demo/grafana.ini /etc/grafana/grafana.ini
COPY grafana/postgres_datasource.yml /etc/grafana/provisioning/datasources/pg_ds.yml
COPY grafana/postgres_dashboard.yml /etc/grafana/provisioning/dashboards/pg_db.yml
ADD grafana/postgres/v10/ /var/lib/grafana/dashboards/


# Set up supervisord [https://docs.docker.com/engine/admin/using_supervisord/]
COPY docker/demo/supervisord.conf /etc/supervisor/
COPY docker/bootstrap/init_persistent_config.sh \
     docker/bootstrap/init_supervisor.sh \
     docker/bootstrap/init_test_monitored_db.sh\
        /pgwatch3/bootstrap/

# Set up PostgreSQL initialization scripts
RUN  mkdir -p /docker-entrypoint-initdb.d
COPY docker/bootstrap/create_role_db.sql \
     docker/bootstrap/init_dbs.sh \
        /docker-entrypoint-initdb.d/

# Apply PostgreSQL configuration
RUN  echo "include = 'pgwatch_postgresql.conf'" >> /etc/postgresql/postgresql.conf
COPY docker/demo/postgresql.conf /etc/postgresql/pgwatch_postgresql.conf

ENV PW3_CONFIG postgresql://pgwatch3:pgwatch3admin@localhost:5432/pgwatch3
ENV PW3_SINK postgresql://pgwatch3:pgwatch3admin@localhost:5432/pgwatch3_metrics
ENV PW3_AES_GCM_KEYPHRASE_FILE /pgwatch3/persistent-config/default-password-encryption-key.txt
ENV POSTGRES_PASSWORD=mysecretpassword

# Admin UI for configuring servers to be monitored
EXPOSE 8080
# Postgres DB holding the pgwatch3 config DB / metrics
EXPOSE 5432
# Grafana UI
EXPOSE 3000
# Prometheus scraping port
EXPOSE 9187

### Volumes for easier updating to newer to newer pgwatch3 containers
### Backwards compatibility is not 100% guaranteed so a backup
### using traditional means is still recommended before updating - see "Updating to a newer Docker version" from README

VOLUME /pgwatch3/persistent-config
VOLUME /var/lib/postgresql
VOLUME /var/lib/grafana


CMD ["/usr/bin/supervisord", "--configuration=/etc/supervisor/supervisord.conf", "--nodaemon"]
ARG grafana_version=latest
ARG grafana_image=grafana-enterprise

FROM golang:1.20 AS backend
WORKDIR /mage
RUN git clone https://github.com/magefile/mage
WORKDIR /mage/mage
RUN go run bootstrap.go
WORKDIR /app/grafana-infinity-datasource/backend
COPY pkg/ ./pkg
COPY Magefile.go ./
COPY go.mod ./
COPY go.sum ./
COPY ./src/plugin.json ./src/
RUN go mod download
RUN mage -v

FROM node:16 as frontend
WORKDIR /app/grafana-infinity-datasource/frontend
COPY README.md ./
COPY CHANGELOG.md ./
COPY LICENSE ./
COPY .npmrc ./
COPY package.json ./
RUN npm install @mercury/client-adapter 
COPY yarn.lock ./
COPY tsconfig.json ./
COPY jest.config.js ./
COPY .eslintrc.ts ./
COPY .prettierrc.js ./
COPY cspell.config.json ./
COPY src/ ./src/
RUN yarn install --frozen-lockfile
RUN yarn dev



FROM grafana/grafana-enterprise:main
WORKDIR /var/lib/grafana/plugins/yesoreyeram-infinity-datasource
COPY --from=backend /app/grafana-infinity-datasource/backend/dist ./dist/
COPY --from=frontend /app/grafana-infinity-datasource/frontend/dist ./dist/
ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS yesoreyeram-infinity-datasource
ADD ./try/dashboards /dashboards/
ADD ./try/provisioning/datasources/default.yml /etc/grafana/provisioning/datasources/default.yml
ADD ./try/provisioning/dashboards/default.yml /etc/grafana/provisioning/dashboards/default.yml
ENTRYPOINT [ "/run.sh" ]

FROM grafana/${grafana_image}:${grafana_version}
WORKDIR /var/lib/grafana/plugins/yesoreyeram-infinity-datasource
COPY --from=backend /app/grafana-infinity-datasource/backend/dist ./dist/
COPY --from=frontend /app/grafana-infinity-datasource/frontend/dist ./dist/
ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS yesoreyeram-infinity-datasource
ADD ./try/dashboards /dashboards/
ADD ./try/provisioning/datasources/default.yml /etc/grafana/provisioning/datasources/default.yml
ADD ./try/provisioning/dashboards/default.yml /etc/grafana/provisioning/dashboards/default.yml
ENTRYPOINT [ "/run.sh" ]

# Make it as simple as possible to access the grafana instance for development purposes
# Do NOT enable these settings in a public facing / production grafana instance

# Set development mode so plugins can be loaded without the need to sign


# Inject livereload script into grafana index.html
USER root
RUN sed -i 's/<\/body><\/html>/<script src=\"http:\/\/localhost:35729\/livereload.js\"><\/script><\/body><\/html>/g' /usr/share/grafana/public/views/index.html
### This file is using multi-stage builds https://docs.docker.com/develop/develop-images/multistage-build/
### It requires 17.05 or higher to be run

########################################################################
### builder stage : Build the golang application in src folder
FROM golang:1.19-alpine as builder

COPY . /src
WORKDIR /src
RUN go build -o bin/main cmd/ability/main.go
########################################################################


########################################################################
### app stage : Contains only binary and config and exposes the command
FROM alpine:latest as app
LABEL maintainer="celian.garcia1@gmail.com"

# Some arguments used for labelling
ARG BUILD_DATE
ARG VCS_REF
ARG BUILD_VERSION
ARG GITLAB_TOKEN
ARG PROJECT_NAME
ARG MODULE_NAME
ARG MODULE_DESCRIPTION
ARG DOCKER_IMAGE

# Labels.
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.name="$PROJECT_NAME::$MODULE_NAME"
LABEL org.label-schema.description=$MODULE_DESCRIPTION
LABEL org.label-schema.url="https://www.$PROJECT_NAME.com/"
LABEL org.label-schema.vcs-url="https://github.com/$PROJECT_NAME/$MODULE_NAME"
LABEL org.label-schema.vcs-ref=$VCS_REF
LABEL org.label-schema.version=$BUILD_VERSION
LABEL org.label-schema.docker.cmd="docker run -it $DOCKER_IMAGE:$BUILD_VERSION"

# Two files are necessary from the build stage : the configuration and the binary
ENV CONFIGURATION_PATH=/etc/ability/config.toml
ENV BINARY_PATH=/bin/ability

COPY --from=builder /src/config.toml ${CONFIGURATION_PATH}
COPY --from=builder /src/bin/main $BINARY_PATH

# Build the main command
CMD .$BINARY_PATH
########################################################################

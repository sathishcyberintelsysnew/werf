ARG STAPEL_SHELL_IMAGE_NAME="no_such_image"


FROM ${STAPEL_SHELL_IMAGE_NAME} AS builder

ADD src/file* /app/added/
ADD "https://github.com/octocat/Hello-World/tarball/7fd1a60b01f91b314f59955a4e4d4e80d8edf11d" /helloworld.tgz
COPY src/file* /app/copied/


FROM ubuntu:22.04 AS result

ARG CHANGED_ARG="should_be_changed"
ARG BASE_STAPEL_SHELL_IMAGE_NAME="no_such_image"
ARG BASE_STAPEL_SHELL_IMAGE_ID="no_such_image"
ARG BASE_STAPEL_SHELL_IMAGE_REPO="no_such_image"
ARG BASE_STAPEL_SHELL_IMAGE_TAG="no_such_image"

ENV COMPOSED_ENV="env-${CHANGED_ARG}"
LABEL COMPOSED_LABEL="label-${CHANGED_ARG}"
MAINTAINER "maintainer"

SHELL ["/bin/sh", "-c"]
USER 0:0
WORKDIR /

COPY --from=builder /app /app
COPY --from=builder /helloworld.tgz /helloworld.tgz

RUN touch /created-by-run-state0
RUN mkdir -p /volume10/should-exist-in-volume

RUN echo ${BASE_STAPEL_SHELL_IMAGE_NAME} >> base_image_data.txt
RUN echo ${BASE_STAPEL_SHELL_IMAGE_ID} >> base_image_data.txt
RUN echo ${BASE_STAPEL_SHELL_IMAGE_REPO} >> base_image_data.txt
RUN echo ${BASE_STAPEL_SHELL_IMAGE_TAG} >> base_image_data.txt

# Complex ARG and ENV usage
ENV TEST_ENV=$PATH:test-env1
ENV PATH=$TEST_ENV:custom-path
ARG TEST_ARG=${TEST_ENV}:${PATH}
ARG TEST_ARG2=${TEST_ARG}:test-arg-2
ARG ANOTHER_ARG=$PATH
ENV PATH=$PATH:/opt/bin
RUN echo ${PATH} > /PATH
RUN echo ${ANOTHER_ARG} > /ANOTHER_ARG
RUN echo ${TEST_ARG} > /TEST_ARG
RUN echo ${TEST_ARG2} > /TEST_ARG2

COPY script.sh /app/
ARG USERNAME
ARG PASSWORD
RUN /app/script.sh

ENTRYPOINT ["sh", "-ec"]
CMD ["tail -f /dev/null"]
VOLUME /volume10
ONBUILD RUN echo onbuild
STOPSIGNAL SIGTERM
HEALTHCHECK CMD echo healthcheck10
EXPOSE 8000/tcp

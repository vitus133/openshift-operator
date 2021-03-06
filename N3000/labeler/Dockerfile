# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

FROM golang:1.14 as builder

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o node_labeler main.go

FROM registry.access.redhat.com/ubi8:latest

ARG VERSION
### Required OpenShift Labels
LABEL name="Intel® FPGA discovery container" \
    vendor="Intel Corporation" \
    version=$VERSION \
    release="1" \
    summary="Intel® FPGA discovery container labels kubernetes nodes that include Intel® FPGA accelerator" \
    description="The container tries to detect Intel® FPGA accelerators and label the nodes that contain them \
and remove the label for nodes that do not."

RUN mkdir -p /usr/share/misc/ && \
    cd /usr/share/misc/ && \
    curl https://pci-ids.ucw.cz/v2.2/pci.ids.gz --output pci.ids.gz && \
    gzip -d pci.ids.gz && \
    rm -rfv ./pci.ids.gz

COPY TEMP_LICENSE_COPY /licenses/LICENSE
USER 1001
WORKDIR /labeler-workspace
COPY --from=builder /workspace/node_labeler .

ENTRYPOINT ["bash", "-c", "/labeler-workspace/node_labeler && sleep infinity"]

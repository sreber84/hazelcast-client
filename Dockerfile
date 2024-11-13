# Build Stage

FROM registry.fedoraproject.org/fedora:latest AS BuildStage

RUN dnf --setopt=install_weak_deps=False install -y golang-bin git

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /main

# Deploy Stage

FROM registry.fedoraproject.org/fedora:latest

WORKDIR /

COPY --from=BuildStage /main /main

CMD [ "/main" ]
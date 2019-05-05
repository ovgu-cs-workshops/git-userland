FROM golang:1.12-alpine as builder
RUN apk update && apk add build-base git
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -ldflags '-linkmode=external "-extldflags=-static"'
RUN cp git-userland /bin

FROM debian:stretch
ARG APP_VERSION
ARG APP_NAME
LABEL version "${APP_VERSION}"
LABEL service "${APP_NAME}"
LABEL vendor "EmbeddedEnterprises"
LABEL product "robµlab"
LABEL maintainers "Martin Koppehel <mkoppehel@embedded.enterprises>"
RUN echo "en_US.UTF-8 UTF-8" > /etc/locale.gen
RUN apt update && apt install -y locales man-db
RUN locale-gen
RUN groupadd -g 1000 user
RUN useradd -d /home/user -g 1000 -u 1000 -o -m -s /bin/zsh user
RUN apt update && apt install -y zsh git tmux vim emacs-nox tig nano less patch git-man gnupg2
RUN apt install curl -y && su user -c 'sh -c "$(curl -fsSL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"' && sed -i s/robbyrussel/bira/g /home/user/.zshrc && apt remove curl -y && apt autoremove -y

COPY --from=builder /bin/git-userland /bin/git-userland

WORKDIR /home/user
ADD ./setup-userland.sh /setup-userland.sh
ADD ./start-userland.sh /start-userland.sh

ENTRYPOINT ["sh"]
CMD ["/start-userland.sh"]
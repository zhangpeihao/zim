FROM alpine

MAINTAINER Zhang Peihao <zhangpeihao@gmail.com>

ADD release/linux/amd64/zim /zim

ENTRYPOINT ["/zim"]

EXPOSE 8870 8871 8872 8873

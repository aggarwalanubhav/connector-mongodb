FROM ubuntu

RUN apt-get update
RUN apt-get upgrade
RUN apt-get install -y golang mongodb git vim

WORKDIR /go/src/app
COPY . .

RUN git clone https://github.com/storj-thirdparty/connector-mongodb.git
RUN cd connector-mongodb
RUN go build

ENTRYPOINT [ "mongo" ]
CMD [ "use storjdb \
db.collection.insert({name: 'Adhyan', age: 21})"]

WORKDIR /go/src/app/connector-mongodb/cmd
RUN go test -v
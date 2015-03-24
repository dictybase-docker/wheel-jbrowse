FROM dictybase/bioperl-zmq-golang
MAINTAINER Siddhartha Basu <siddhartha-basu@northwestern.edu>
ENV PATH /go/bin:$PATH
RUN mkdir -p /usr/src/jbrowse
COPY runner/ /usr/src/jbrowse/
RUN cd /usr/src/jbrowse && cpanm -n --installdeps .
CMD ["app"]


FROM nginx
COPY hipaa/ /usr/local/bin/hipaa
COPY entrypoint.sh /bin/entrypoint
RUN chmod +x /bin/entrypoint
RUN apt-get update -y && apt-get install ruby procps vim ruby-dev g++ make -y && gem install rest-client
CMD ["/bin/entrypoint"]
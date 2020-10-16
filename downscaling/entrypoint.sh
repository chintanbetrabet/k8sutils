#!/bin/bash -l
set -x

handle_sigterm() {
  echo "Caught SIGTERM..."
  ruby /usr/local/bin/hipaa/pre_shutdown.rb
  ruby /bin/graceful_shutdown.rb
  ruby /usr/local/bin/hipaa/post_shutdown.rb
}
#0 Trap signals for graceful termination
trap handle_sigterm SIGTERM

ruby /usr/local/bin/hipaa/pre_start.rb

sleep 600

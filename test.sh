#!/bin/sh

set -e

function with_remote()
{
    echo "with remote"
    traceid=$(uuidgen | tr -d '-')

    echo $traceid

    # curl with trace context
    curl -H "traceparent: 00-$traceid-0000000000000001-01" \
         -H "tracestate: foo=bar" \
         -H "user-agent: curl/7.64.1" \
         -H "accept: */*" \
         -X GET \
         http://127.0.0.1:1323/

}


function without_remote()
{
    echo "without remote context"
    # curl with trace context
    curl -X GET \
         -H 'Content-Type: application/json' \
         http://127.0.0.1:1323/
}


case "$@" in
    remote)
      with_remote
      ;;
    *)
      without_remote
      ;;
esac
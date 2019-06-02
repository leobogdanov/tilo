#!/bin/bash

flyway -configFiles=/app/flyway.conf repair \
&& flyway  -configFiles=/app/flyway.conf migrate
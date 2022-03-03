#!/usr/bin/env sh

export GOMAXPROC=1
export AGENTD_LOG_FILE=commanderd.log

nohup ./agentd >> ${AGENTD_LOG_FILE} &
tail -f ${AGENTD_LOG_FILE} &

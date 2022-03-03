#!/usr/bin/env sh

export GOMAXPROC=1
export COMMANDERD_CONFIG_FILE=commanderd.conf.json
export COMMANDERD_LOG_FILE=commanderd.log
export AGENTD_LOG_FILE=commanderd.log

nohup ./commanderd -conf ${COMMANDERD_CONFIG_FILE} >> ${COMMANDERD_LOG_FILE} &
tail -f ${COMMANDERD_LOG_FILE} &

nohup ./agentd > ${AGENTD_LOG_FILE} &
tail -f ${AGENTD_LOG_FILE} &

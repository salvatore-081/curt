#!/bin/bash
params=()
[[ $PORT ]] && params+=(-PORT $PORT)
[[ $LOG_LEVEL ]] && params+=(-LOG_LEVEL $LOG_LEVEL)
[[ $API_KEY ]] && params+=(-API_KEY $API_KEY)

/app/curt ${params[@]}
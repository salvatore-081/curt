#!/bin/bash
params=()
[[ $PORT ]] && params+=(-PORT $PORT)
[[ $LOG_LEVEL ]] && params+=(-LOG_LEVEL $LOG_LEVEL)
[[ $X_API_KEY ]] && params+=(-X_API_KEY $X_API_KEY)
[[ $HOST ]] && params+=(-HOST $HOST)

/app/curt ${params[@]}
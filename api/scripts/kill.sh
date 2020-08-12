#!/bin/bash

ps -ef | grep dlv | grep -v grep | awk '{ print $2 }' | xargs kill
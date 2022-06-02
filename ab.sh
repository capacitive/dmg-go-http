#!/bin/bash
ab -n 100000 -c 20000 "127.0.0.1:8000/increment"
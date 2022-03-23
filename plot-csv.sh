#!/bin/bash

echo "* Plotting '$1' using GNUPlot"

gnuplot -p <<EOF
 set datafile separator ","
 plot "$1" using 0:1 with lines
EOF

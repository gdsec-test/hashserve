#!/bin/bash
################################################################################
# 
# IMPORTANT NOTICE
# ================
# 
#   Copyright (C) 2016, Microsoft Corporation
#   All Rights Reserved.
# 
#   THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS STRICTLY UNDER THE TERMS
#   OF A SEPARATE LICENSE AGREEMENT.  NO SOURCE CODE DISTRIBUTION OR SOURCE 
#   CODE DISCLOSURE RIGHTS ARE PROVIDED WITH RESPECT TO THIS SOFTWARE OR ANY 
#   DERIVATIVE WORKS THEREOF.  USE AND NON-SOURCE CODE REDISTRIBUTION OF THE 
#   SOFTWARE IS PERMITTED ONLY UNDER THE TERMS OF SUCH LICENSE AGREEMENT.  
#   THIS SOFTWARE AND ANY WORKS/MODIFICATIONS/TRANSLATIONS DERIVED FROM THIS 
#   SOFTWARE ARE COVERED BY SUCH LICENSE AGREEMENT AND MUST ALSO CONTAIN THIS 
#   NOTICE.  THIS SOFTWARE IS CONFIDENTIAL AND SENSITIVE AND MUST BE SECURED 
#   WITH LIMITED ACCESS PURSUANT TO THE TERMS OF SUCH LICENSE AGREEMENT.
# 
#   THE SOFTWARE IS PROVIDED "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, 
#   INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY 
#   AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
#   COPYRIGHT HOLDER BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, 
#   EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, 
#   PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
#   OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, 
#   WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR 
#   OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF 
#   ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
#   
#   Project:       PhotoDNA Robust Hashing
#   File:          TestGenerate.sh
#   Description:   Tool to test GenerateHashes sample
#
#   History:   2015.11.20   adrian chandley  Created test. 
#  
################################################################################

#
# TestGenerate [executable to test]
#


#
# Calculate [image file] [condition] [value]
#
#   Uses the executable to generate a hash for the given image
#   The resulting hash will be compared to the the base test hash using the given condition
#
#   If the condition is not met (e.g. gt 1000), the executable has failed the test
#
Calculate() {
dist=0
pushd "$SCRIPTDIR" > /dev/null

# run executable
pcmd=$PROGRAM
case "${PROGRAM##*\.}" in
  py)     pcmd="python $PROGRAM";;
  java)   pcmd="java -cp $PROGRAMDIR $(basename "$PROGRAM" .java)";;
  class)  pcmd="java -cp $PROGRAMDIR $(basename "$PROGRAM" .class)";;
esac
Hash=$($pcmd "$1")

HASH=(${Hash//,/ })
for i in {1..144}
  do
    diff=$[${BASE[$i]}-${HASH[$i]}]
    dist=$[$dist+ ($diff * $diff)]
 done
popd > /dev/null
if [ $dist -$2 $3 ]; then 
  echo FAIL: Distance for $1: $dist
  return 2
else 
  echo PASS: Distance for $1: $dist
  return 0
fi
}

#
# TestGenerate [executable to test]
#

#get the full path of the executable
PROGRAMDIR=$( cd "$( dirname "$1" )" && pwd )
PROGRAM=$PROGRAMDIR/${1##*/}
if [ ! -f "$PROGRAM" ]; then
  echo "ERROR: Program not found"
  exit 1
fi

# Get the executables date
fd=""
case "$OSTYPE" in
  darwin*)  fd="(File time: "$(date -r $(stat -f "%m" $1) '+%Y.%m.%d %H:%M')")" ;; 
  linux*)   fd="(File time: "$(date -d @$(stat -c "%Y" $1) '+%Y.%m.%d %H:%M')")" ;;
esac
echo Testing: $1 $fd

SCRIPTDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

Base=B,5,70,1,78,27,18,8,58,14,24,16,55,15,28,63,51,32,38,57,68,42,14,48,102,1,53,214,42,6,34,209,64,4,22,178,81,6,15,146,76,20,10,127,61,45,0,128,57,5,47,255,6,1,57,255,6,0,52,255,13,21,28,255,13,51,14,255,5,77,4,255,14,13,41,17,92,24,33,29,88,29,52,71,31,60,18,91,26,75,2,69,12,45,17,63,32,28,86,67,44,124,18,62,46,56,97,75,69,64,46,58,108,101,17,32,66,54,22,36,65,62,53,88,31,32,80,44,73,81,59,42,227,81,63,133,101,128,43,85,68,88,15,183,21
BASE=(${Base//,/ })

_res=0
Calculate "Test01.jpg" gt 1000
_res=$[$_res + $?]
Calculate "Test02.jpg" lt 1000
_res=$[$_res + $?]
Calculate "Test03.jpg" gt 1000
_res=$[$_res + $?]
Calculate "Test04.png" gt 1000
_res=$[$_res + $?]

exit $_res

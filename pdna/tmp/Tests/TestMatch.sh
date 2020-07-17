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
#   File:          TestMatch.sh
#   Description:   Tool to test MatchHashes sample
#
#   History:   2015.11.20   adrian chandley  Created test. 
#  
################################################################################

#
# TestMatch [executable to test]
#


#
# Match [test candidate hashes] [test hash database] [Image name]
#
#   Uses the executable to compare the candidate hashes against the hash database
# 
#   If a match is not found for [Image Name], the executable has failed the test 
#
Match() {
dist=-1
match=""
pushd "$SCRIPTDIR" > /dev/null

# run executable
pcmd=$PROGRAM
case "${PROGRAM##*\.}" in
  py)     pcmd="python $PROGRAM";;
  java)   pcmd="java -cp $PROGRAMDIR $(basename "$PROGRAM" .java)";;
  class)  pcmd="java -cp $PROGRAMDIR $(basename "$PROGRAM" .class)";;
esac
output=$($pcmd "$1" "$2" "" "test" 2>/dev/null)

SAVEIFS="${IFS}"
IFS=$'\n'
for x in ${output} ; do
  if [ "${x:0:5}" == "Test0" ]; then
    resp=(${x//,/${IFS}})
    if [ "${resp[1]}" -lt "1000" ]; then
      dist=${resp[1]}
      match=${resp[0]}
    fi
  fi
done
IFS="${SAVEIFS}"

if [ $dist -eq -1 ]; then 
  echo FAIL: Match incorrect \($3\)
  return 2
else 
  echo PASS: Correct match found \($3\)
  return 0
fi
popd > /dev/null
}

#
# TestMatch [executable to test]
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

_res=0
Match Test1.csv Test.csv Test03.jpg
_res=$[$_res + $?]
Match Test2.csv Test.csv Test04.png
_res=$[$_res + $?]

exit $_res

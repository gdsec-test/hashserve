################################################################################
# 
# IMPORTANT NOTICE
# ================
# 
#   Copyright (C) 2016, 2017, Microsoft Corporation
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
#   File:          GNUmakefile
#   Description:   Makefile for C++ samples (make - linux osx)
#
#   History:  2016.07.26   adrian chandley  Created
#             2016.10.31   adrian chandley  Added java support
#
################################################################################
#
#  Usage: make [-B]
#
################################################################################

LibVersion=1.72

UNAME_S := $(shell uname -s)
PROCESSOR := $(shell uname -m)
LONG_BIT := $(shell getconf LONG_BIT)
ifeq ($(LONG_BIT),64)
  platform=x64
  ifneq (,$(findstring arm,$(PROCESSOR)))
    platform=arm64
  endif
  ifneq (,$(findstring aarch,$(PROCESSOR)))
    platform=arm64
  endif
else
  platform=x86
  ifneq (,$(findstring arm,$(PROCESSOR)))
    platform=arm32
  endif
  ifneq (,$(findstring aarch,$(PROCESSOR)))
    platform=arm32
  endif
endif

JAVAC := $(shell which javac)

CC=g++ -Wall -O3 -DENVIRONMENT$(platform)

cppbin=bin/cpp
cppsrc=src/cpp
javabin=bin/java
javasrc=src/java
pybin=bin/python
testdir=tmp/Tests

all: c++ python java complete simple
cpp: c++
simple: $(platform) $(platform)/SimpleExample
c++: $(platform) $(platform)/GenerateHashes $(platform)/MatchHashes $(platform)/ScanVideo
python: PythonTest.txt
java: GenerateHashes.class MatchHashes.class 

$(platform)/GenerateHashes: $(cppsrc)/GenerateHashes.cpp $(cppsrc)/RobustHash.h $(cppsrc)/RobustHashCalls.hpp
	@echo "Building GenerateHashes (requires opencv.org libraries to be installed)"
	@$(CC) -I /opt/local/include/ -o $(cppbin)/$(platform)/GenerateHashes $(cppsrc)/GenerateHashes.cpp `pkg-config --libs opencv` -ldl
	@chmod u+x $(testdir)/TestGenerate.sh
	@$(testdir)/TestGenerate.sh $(cppbin)/$(platform)/GenerateHashes

$(platform)/SimpleExample: $(cppsrc)/SimpleExample.cpp $(cppsrc)/RobustHash.h $(cppsrc)/RobustHashCalls.hpp
	@echo "Building SimpleExample"
	@$(CC) -I /opt/local/include/ -o $(cppbin)/$(platform)/SimpleExample $(cppsrc)/SimpleExample.cpp -ldl
	@if [ ! -z "$(JAVAC)" ]; then javac -d $(javabin) $(javasrc)/SimpleExample.java $(javasrc)/PhotoDNA.java;fi

$(platform)/MatchHashes: $(cppsrc)/MatchHashes.cpp $(cppsrc)/RobustHashTools.cpp $(cppsrc)/RobustHash.h $(cppsrc)/RobustHashCalls.h
	@echo "Building MatchHashes"
	@$(CC) -o $(cppbin)/$(platform)/MatchHashes $(cppsrc)/MatchHashes.cpp $(cppsrc)/RobustHashTools.cpp 
	@chmod u+x $(testdir)/TestMatch.sh
	@$(testdir)/TestMatch.sh $(cppbin)/$(platform)/MatchHashes

GenerateHashes.class: $(javasrc)/GenerateHashes.java $(javasrc)/PhotoDNA.java
	@if [ ! -z "$(JAVAC)" ]; then echo "Building GenerateHashes (Java)"; else echo "Java complier not found";fi
	@if [ ! -z "$(JAVAC)" ]; then javac -d $(javabin) $(javasrc)/GenerateHashes.java $(javasrc)/PhotoDNA.java;fi
	@if [ ! -z "$(JAVAC)" ]; then $(testdir)/TestGenerate.sh $(javabin)/GenerateHashes.class;fi

MatchHashes.class: $(javasrc)/MatchHashes.java
	@if [ ! -z "$(JAVAC)" ]; then echo "Building MatchHashes (Java)"; else echo "Java complier not found";fi
	@if [ ! -z "$(JAVAC)" ]; then javac -d $(javabin) $(javasrc)/MatchHashes.java;fi
	@if [ ! -z "$(JAVAC)" ]; then $(testdir)/TestMatch.sh $(javabin)/MatchHashes.class;fi

$(platform)/ScanVideo: $(cppsrc)/ScanVideo.cpp $(cppsrc)/RobustHash.h $(cppsrc)/RobustHashCalls.hpp $(cppsrc)/Hashes.hpp
	@echo "Building ScanVideo"
	@$(CC) -o $(cppbin)/$(platform)/ScanVideo $(cppsrc)/ScanVideo.cpp -ldl

PythonTest.txt: $(pybin)/GenerateHashes.py
	@echo "Running Python Tests"
	@chmod u+x $(testdir)/*.sh
	@$(testdir)/TestGenerate.sh $(pybin)/GenerateHashes.py > GenerateTest.txt
	@cat GenerateTest.txt
	@$(testdir)/TestMatch.sh $(pybin)/MatchHashes.py > MatchTest.txt
	@cat GenerateTest.txt > PythonTest.txt
	@cat MatchTest.txt
	@cat MatchTest.txt >> PythonTest.txt
	@rm GenerateTest.txt
	@rm MatchTest.txt

$(platform):
	@if [ ! -d "$(cppbin)/$(platform)" ]; then mkdir $(cppbin)/$(platform);fi
	@cp /usr/local/lib/PhotoDNA$(platform)$(suffix).so.$(LibVersion) $(cppbin)/$(platform)

complete:
	@echo "Build completed $(COMPLETE)"

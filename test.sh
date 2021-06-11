	COMPLETE_COUNT=$(/var/github-actions/_work/_tool/kubectl/1.15.0/x64/kubectl --context dev-dcu -n abuse-api-dev get pods | grep -i shadowfax-integration-test | grep -i completed | wc -l | xargs );\
	if [ "$$COMPLETE_COUNT" = "1" ]; then \
    echo "Integration test passed successfully"; \
	else \
	  echo "Integration tests failed!!!!! Revert changes manually.";
	   exit 1
	fi;
.PHONY: run stress

run:
	$(POWERSHELL) -File ./run.ps1
	@echo "AgentMesh is live at http://localhost:5173"
ifeq ($(OS),Windows_NT)
	@cmd /c start http://localhost:5173
else
	@uname | grep Darwin > /dev/null && open http://localhost:5173 || xdg-open http://localhost:5173
endif

stress:
	go run cmd/stress_test/main.go -count=500 -concurrency=20

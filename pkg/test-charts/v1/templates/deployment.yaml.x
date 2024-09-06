apiVersion: apps/v1
kind: Deployment
metadata:
  name: this-chart
  namespace: {{ .Release.Namespace | quote }}
spec:
  template:
    spec:
      containers:
        - name: {{ .Chart.Name }}
          env:
            - name: "TargetEnvironment"
              value: {{ join "," .Values.agent.targetEnvironments | quote }}

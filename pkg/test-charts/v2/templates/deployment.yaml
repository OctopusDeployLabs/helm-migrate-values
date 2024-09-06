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
            - name: "TargetEnvironments"
              value: [{{ join "," .Values.agent.deploymentTarget.initial.environments | quote }}]
  selector:
    matchLabels:
      app: {{ .Chart.Name }}
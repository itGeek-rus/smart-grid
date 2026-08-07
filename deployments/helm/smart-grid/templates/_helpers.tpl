{{- define "smart-grid.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "smart-grid.labels" -}}
app.kubernetes.io/name: {{ include "smart-grid.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

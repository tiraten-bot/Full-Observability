{{/*
Expand the name of the chart.
*/}}
{{- define "full-observability.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "full-observability.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "full-observability.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "full-observability.labels" -}}
helm.sh/chart: {{ include "full-observability.chart" . }}
{{ include "full-observability.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "full-observability.selectorLabels" -}}
app.kubernetes.io/name: {{ include "full-observability.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "full-observability.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "full-observability.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database connection string
*/}}
{{- define "full-observability.dbHost" -}}
{{- printf "postgres.%s.svc.cluster.local" .Values.global.namespace }}
{{- end }}

{{/*
Redis connection string
*/}}
{{- define "full-observability.redisHost" -}}
{{- printf "redis.%s.svc.cluster.local" .Values.global.namespace }}
{{- end }}

{{/*
Kafka bootstrap servers
*/}}
{{- define "full-observability.kafkaBootstrap" -}}
{{- printf "kafka.%s.svc.cluster.local:29092" .Values.global.namespace }}
{{- end }}

{{/*
Jaeger endpoint
*/}}
{{- define "full-observability.jaegerEndpoint" -}}
{{- printf "http://jaeger.%s.svc.cluster.local:14268/api/traces" .Values.global.namespace }}
{{- end }}

{{/*
Prometheus URL
*/}}
{{- define "full-observability.prometheusURL" -}}
{{- printf "http://prometheus.%s.svc.cluster.local:9090" .Values.global.namespace }}
{{- end }}


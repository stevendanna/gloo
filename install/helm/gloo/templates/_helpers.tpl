{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}


{{- define "gloo.roleKind" -}}
{{- if .Values.global.glooRbac.namespaced -}}
Role
{{- else -}}
ClusterRole
{{- end -}}
{{- end -}}

{{- define "gloo.rbacNameSuffix" -}}
{{- if .Values.global.glooRbac.nameSuffix -}}
-{{ .Values.global.glooRbac.nameSuffix }}
{{- else if not .Values.global.glooRbac.namespaced -}}
-{{ .Release.Namespace }}
{{- end -}}
{{- end -}}

{{/*
Expand the name of a container image
*/}}
{{- define "gloo.image" -}}
{{ .registry }}/{{ .repository }}:{{ .tag }}
{{- end -}}

{{/* This value makes its way into k8s labels, so if the implementation changes,
     make sure it's compatible with label values. Expects its root context to
     have an "installationId" key that will be unique through the installation. */}}
{{- define "gloo.installationId" -}}
{{- if not .installationId -}}
{{- $_ := set . "installationId" (randAlphaNum 20) -}}
{{- end -}}
{{ .installationId }}
{{- end -}}

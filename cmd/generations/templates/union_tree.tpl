union{{ with .FamilyID }}{{ if . }}[id={{ . }}]{{ end }}{{ end }} {
    {{ .Parent }}
    {{ .Children }}
}
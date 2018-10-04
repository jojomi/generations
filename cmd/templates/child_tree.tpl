child{{ with .FamilyID }}{{ if . }}[id={{ . }}]{{ end }}{{ end }} {
    {{ .G }}
    {{ .Parent }}
    {{ .Children }}
    {{ .Unions }}
}
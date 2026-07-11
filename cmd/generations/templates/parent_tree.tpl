parent{{ with .FamilyID }}{{ if . }}[id={{ . }}]{{ end }}{{ end }} {
    {{ .SiblingsOlder }}
    {{ .G }}
    {{ .SiblingsYounger }}
    {{ .Parents }}
}
{{- .Options.NodeType -}}[%
  {{ if not .Options.HideID -}}
    {{ if .Person.GetID -}}
      id={{ .Person.GetID }},
    {{ else }}
      {{ if .Person.GetUUID -}}
        id={{ .Person.GetUUID }},
      {{- end }}%
    {{- end }}%
  {{- end }}%

  {{- $hideList := $.Options.HideAttributes }}
  {{ if eq (len $hideList) 1 -}}
  {{ if eq (index $hideList 0) "all" }}%
  {{ $hideList = .Person.GetAttributes }}%
  {{- end }}%
  {{- end }}%
  {{ $attributes := getFilteredStringSlice .Person.GetAttributes $hideList }}%
  ,%
  {{ if .Person.GetDeadState }}%
    dead,
  {{- end }}%
]{
  {{ with .Person.GetUUID }}
    {{ if . }}
      uuid={{ . }},
    {{ end }}
  {{ end }}

  {{ $gender := .Person.GetGender }}
  {{ if and (not $gender.IsUnknown) (not .Options.HideGender) }}
    {{ with $gender }}
      sex = {{ . }},
    {{ end }}
  {{ end }}

  {{ $name := .Person.GetName }}
  {{ if and (not $name.Empty) (not .Options.HideName) }}
    {{ with .Person.GetName }}
      name = {%
        {{ with $name.Title }}{{ if . }}\titlename{ {{- . -}} }\ %{{ end }}{{ end }}
        {{ if $name.Used }}
          {{ $matched := false }}
          {{ range .First }}
            {{ if ne . $name.Used }}
              {{ if and . (not $.Options.HideMiddleNames) }}
                \middlename{ {{- . -}} }\ %
              {{ end }}
            {{ else }}
              {{ $matched = true }}
              \pref{ {{- . -}} }\ %
            {{ end }}
          {{ end }}
          {{ if and (not $matched) (ne $name.Used "?") }}
          \pref{ {{- $name.Used -}} }\ %
          {{ end }}

        {{ else }}
          {{ range $i, $name := .First }}
            {{ if ne $i 0 }}
              {{ if and . (not $.Options.HideMiddleNames) }}
                \middlename{ {{- $name -}} }\ %
              {{ end }}
            {{ else }}
              \pref{ {{- $name -}} }\ %
            {{ end }}
          {{ end }}
        {{ end }}

      {{ with .Nick }}
      {{ if . }}
      \nick{ {{- . -}} }\ %
      {{ end }}
      {{ end }}


      {{ if or (eq $.Options.LastnamePolicy.String "CurrentAndBirth") (eq $.Options.LastnamePolicy.String "Current") }}
      {{ if .Last }}
        \surn{ {{- .Last -}} }%
      {{- end }}
      {{ end }}

      {{ if eq $.Options.LastnamePolicy.String "CurrentAndBirth" }}
      {{- if .Birth -}}
        \surnbirth{ {{- .Birth -}} }\ %
      {{ end }}
      {{ end }}

      {{ if eq $.Options.LastnamePolicy.String "Birth" }}
      {{- if .Birth -}}
        \surn{ {{- .Birth -}} }\ %
      {{ else }}
        \surn{ {{- .Last -}} }\ %
      {{ end }}
      {{ end }}


      {{ with .Alias }}
      {{ if . }}
      \\ \alias{ {{- . -}} }\ %
      {{ end }}
      {{ end }}

      },
    {{ end }}
  {{ end }}


  {{ $birth := .Person.GetBirth }}
  {{ if and (not $birth.Empty) (not .Options.HideBirth) }}
    {{ with $birth }}
      {{ if and (not $.Options.HidePlaces) (ne .Place "") }}
      birth = { {{- if .Date }}{{ .Date }}{{ else }}-{{ end -}} }{ {{- .Place -}} },
      {{ else }}
      birth- = { {{- .Date -}} },
      {{ end }}

      {{ $age := $.Person.GetAge $.Options.Date }}
      {{ with $age }}
        {{ if and (ne . -1) ($.Options.ShowAge) ($.Person.GetDeath.Empty) }}
          age = { {{- . -}} },
        {{ end }}
      {{ end }}
    {{ end }}
  {{ end }}

  {{ $baptism := .Person.GetBaptism }}
  {{ if and (not $baptism.Empty) (not .Options.HideBaptism) }}
    {{ with $baptism }}
      {{ if and (not $.Options.HidePlaces) (ne .Place "") }}
      baptism = { {{- if .Date }}{{ .Date }}{{ else }}-{{ end -}} }{ {{- .Place -}} },
      {{ else }}
      baptism- = { {{- .Date -}} },
      {{ end }}
    {{ end }}
  {{ end }}

  {{ $death := .Person.GetDeath }}
  {{ if and (not $death.Empty) (not .Options.HideDeath) }}
    {{ with $death }}
      {{ if and (not $.Options.HidePlaces) (ne .Place "") }}
      death = { {{- if .Date }}{{ .Date }}{{ else }}-{{ end -}} }{ {{- .Place -}} },
      {{ else }}
      death- = { {{- .Date -}} },
      {{ end }}
    {{ end }}
  {{ end }}

  {{ if not $.Options.HideBirth }}
    {{ $deathAge := .Person.GetDeathAge }}
    {{ with $deathAge }}
      {{ if and (ne . -1) (not $.Options.HideDeathAge) }}
        deathage = { {{- . -}} },
      {{ end }}
    {{ end }}
  {{ end }}

  {{ $burial := .Person.GetBurial }}
  {{ if and (not $burial.Empty) (not .Options.HideBurial) }}
    {{ with $burial }}
      {{ if and (not $.Options.HidePlaces) (ne .Place "") }}
      burial = { {{- if .Date }}{{ .Date }}{{ else }}-{{ end -}} }{ {{- .Place -}} },
      {{ else }}
      burial- = { {{- .Date -}} },
      {{ end }}
    {{ end }}
  {{ end }}


  {{ if eq (len .Person.Partners) 1 }}
  {{ with .Person.Partners }}
    {{ range . }}

      {{ with .Engagement }}
        {{ if and (not $.Options.HideEngagement) (not .Empty) }}
          {{ if and (not $.Options.HidePlaces) (ne .Place "") }}
          engagement = { {{- .Date -}} }{ {{- .Place -}} },
          {{ else }}
          engagement- = { {{- .Date -}} },
          {{ end }}
        {{ end }}
      {{ end }}

      {{ template "eventWithAge" (dict "Event" .Marriage "FieldName" "marriage" "AgeFieldName" "marriageage" "HideEvent" $.Options.HideMarriage "HideAge" $.Options.HideMarriageAge "HidePlaces" $.Options.HidePlaces "HideBirth" $.Options.HideBirth "Birth" $birth) }}

      {{ template "eventWithAge" (dict "Event" .MarriageReligious "FieldName" "marriagereligious" "AgeFieldName" "marriageagereligious" "HideEvent" $.Options.HideMarriageReligious "HideAge" $.Options.HideMarriageReligiousAge "HidePlaces" $.Options.HidePlaces "HideBirth" $.Options.HideBirth "Birth" $birth) }}

      {{ with .Divorce }}
        {{ if and (not $.Options.HideDivorce) (not .Empty) }}
          {{ if and (not $.Options.HidePlaces) (ne .Place "") }}
          divorce = { {{- .Date -}} }{ {{- .Place -}} },
          {{ else }}
          divorce- = { {{- .Date -}} },
          {{ end }}
        {{ end }}
      {{ end }}

    {{ end }}
  {{ end }}
  {{ end }}

  {{ if not .Options.HideFloruit }}
  {{ with .Person.Residences }}
    {{ if and . (not $.Options.HidePlaces) }}
      floruit = {-}{
        {{- range $i, $r := . -}}
          {{ if $i }}, {{ end -}}
          {{ $r.City }}
        {{- end -}}
      },
    {{ end }}
  {{ end }}
  {{ end }}

  {{ if not .Options.HideJobs }}
    {{ with .Person.GetJobs }}
      profession = {
        {{- range $i, $j := . -}}
          {{ if $i }}, {{ end -}}
          {{ $j.Type -}}
        {{ end -}}
      },
    {{ end }}
  {{ end }}

  {{ if not .Options.HideImage }}
  {{ with .Person.GetImageFilename }}
    {{ if . }}
      image = { {{- . -}} },
    {{ end }}
  {{ end }}
  {{ end }}

  {{ if not .Options.HideComment }}
  {{ with .Person.GetComments }}
    comment = { 
        {{- range . -}}
        {{ .Text }} {{ end -}}
    },
  {{ end }}
  {{ end }}
}
{{- define "eventWithAge" }}
      {{ with $.Event }}
        {{ if and (not $.HideEvent) (not .Empty) }}
          {{ if and (not $.HidePlaces) (ne .Place "") }}
          {{ $.FieldName }} = { {{- .Date -}} }{ {{- .Place -}} },
          {{ else }}
          {{ $.FieldName }}- = { {{- .Date -}} },
          {{ end }}

          {{ if not $.HideBirth }}
            {{ $age := .GetAgeBegin $.Birth }}
            {{ with $age }}
              {{ if and (ne . -1) (not $.HideAge) }}
                {{ $.AgeFieldName }} = { {{- . -}} },
              {{ end }}
            {{ end }}
          {{ end }}
        {{ end }}
      {{ end }}
{{ end }}
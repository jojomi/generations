{{- .Options.NodeType -}}[%
  {{ if not .Options.HideID }}%
  {{ with .Person.GetID }}%
  id={{ . }},
  {{- end }}%
  {{- end }}%
  {{ if not .Options.HideAttributes }}%
  {{ $attributes := .Person.GetAttributes }}%
  {{ if $attributes }}%
    {{ join $attributes "," }}
  {{- end }}%
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
        {{ if $name.Used }}
          {{ range .First }}
            {{ if ne . $name.Used }}
              {{ if and . (not $.Options.HideMiddleNames) }}
                \middlename{ {{- . -}} }\xspace%
              {{ end }}
            {{ else }}
              \pref{ {{- . -}} }\xspace%
            {{ end }}
          {{ end }}
        {{ else }}
          {{ range $i, $name := .First }}
            {{ if ne $i 0 }}
              {{ if and . (not $.Options.HideMiddleNames) }}
                \middlename{ {{- $name -}} }\xspace%
              {{ end }}
            {{ else }}
              \pref{ {{- $name -}} }\xspace%
            {{ end }}
          {{ end }}
        {{ end }}

      {{ with .Nick }}
      {{ if . }}
      \nick{ {{- . -}} }\xspace%
      {{ end }}
      {{ end }}


      {{ if or (eq $.Options.LastnamePolicy.String "CurrentAndBirth") (eq $.Options.LastnamePolicy.String "Current") }}
      {{ if .Last }}
        \surn{ {{- .Last -}} }\xspace%
      {{- end }}
      {{ end }}

      {{ if eq $.Options.LastnamePolicy.String "CurrentAndBirth" }}
      {{- if .Birth -}}
        \surnbirth{ {{- .Birth -}} }\xspace%
      {{ end }}
      {{ end }}

      {{ if eq $.Options.LastnamePolicy.String "Birth" }}
      {{- if .Birth -}}
        \surn{ {{- .Birth -}} }\xspace%
      {{ else }}
        \surn{ {{- .Last -}} }\xspace%
      {{ end }}
      {{ end }}


      {{ with .Alias }}
      {{ if . }}
      ~-- \alias{ {{- . -}} }\xspace%
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

      {{ with .Marriage }}
        {{ if and (not $.Options.HideMarriage) (not .Empty) }}
          {{ if and (not $.Options.HidePlaces) (ne .Place "") }}
          marriage = { {{- .Date -}} }{ {{- .Place -}} },
          {{ else }}
          marriage- = { {{- .Date -}} },
          {{ end }}
        {{ end }}
      {{ end }}

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
  {{ with .Person.Floruit }}
    {{ if . }}
      floruit- = { {{- . -}} },
    {{ end }}
  {{ end }}
  {{ end }}

  {{ if not .Options.HideJobs }}
  {{ with .Person.GetJobs }}
    {{ if . }}
      profession = { {{- . -}} },
    {{ end }}
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
  {{ with .Person.GetComment }}
    {{ if . }}
      comment = { {{- . -}} },
    {{ end }}
  {{ end }}
  {{ end }}
}
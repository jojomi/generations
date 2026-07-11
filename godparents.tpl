% Godparents
\begin{scope}[on background layer]
  {{ range . -}}
    {{ if .RelatedContactID }}
  \draw[godparent] ({{ encryptID .ContactID "contact" }}) edge[out=240,in=210] node[midway, above, sloped]{} ({{ encryptID .RelatedContactID "contact" }});
    {{ end }}
  {{ end }}
\end{scope}
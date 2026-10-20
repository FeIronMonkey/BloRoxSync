package packageplugin

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const entrySource = `for _, child in script:GetChildren() do
	if child:IsA("ModuleScript") then
		require(child)
	end
end
`

func Package(srcDir, outPath string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	type mod struct {
		name   string
		source string
	}
	var mods []mod
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".luau") && !strings.HasSuffix(name, ".lua") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(srcDir, name))
		if err != nil {
			return err
		}
		stem := strings.TrimSuffix(strings.TrimSuffix(name, ".luau"), ".lua")
		mods = append(mods, mod{name: stem, source: string(body)})
	}
	if len(mods) == 0 {
		return fmt.Errorf("no Luau sources in %s", srcDir)
	}
	sort.Slice(mods, func(i, j int) bool { return mods[i].name < mods[j].name })

	var b strings.Builder
	b.WriteString(`<roblox xmlns:xmime="http://www.w3.org/2005/05/xmlmime" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://www.roblox.com/roblox.xsd" version="4">`)
	b.WriteByte('\n')
	b.WriteString("\t<Item class=\"Script\" referent=\"RBXPLUGIN\">\n")
	b.WriteString("\t\t<Properties>\n")
	b.WriteString("\t\t\t<string name=\"Name\">rbx-sync</string>\n")
	b.WriteString("\t\t\t<ProtectedString name=\"Source\">")
	writeCDATA(&b, entrySource)
	b.WriteString("</ProtectedString>\n")
	b.WriteString("\t\t</Properties>\n")
	for i, m := range mods {
		ref := fmt.Sprintf("RBXMOD%d", i)
		b.WriteString(fmt.Sprintf("\t\t<Item class=\"ModuleScript\" referent=%q>\n", ref))
		b.WriteString("\t\t\t<Properties>\n")
		b.WriteString(fmt.Sprintf("\t\t\t\t<string name=\"Name\">%s</string>\n", xmlEscape(m.name)))
		b.WriteString("\t\t\t\t<ProtectedString name=\"Source\">")
		writeCDATA(&b, m.source)
		b.WriteString("</ProtectedString>\n")
		b.WriteString("\t\t\t</Properties>\n")
		b.WriteString("\t\t</Item>\n")
	}
	b.WriteString("\t</Item>\n")
	b.WriteString("</roblox>\n")

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outPath, []byte(b.String()), 0o644)
}

func writeCDATA(b *strings.Builder, s string) {
	b.WriteString("<![CDATA[")
	b.WriteString(strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>"))
	b.WriteString("]]>")
}

func xmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return r.Replace(s)
}

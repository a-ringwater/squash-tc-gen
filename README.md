# 🧪 squash-tc-gen

CLI Go qui génère un fichier Excel prêt à importer dans Squash TM (format
`TEST_CASES` / `STEPS` / `PARAMETERS` / `DATASETS` / `LINK_REQ_TC`) à partir
d'un fichier YAML décrivant les cas de test d'une story.

Projet réalisé pour automatiser la rédaction de cas de test manuels.

## 💡 Pourquoi

Écrire les cas de test à la main dans l'IHM Squash TM est répétitif, et le
copier-coller depuis un tableau externe est source d'erreurs (chemins,
en-têtes de colonnes, HTML des champs riches). Ce générateur prend une
description structurée en YAML et produit directement le fichier d'import
au bon format.

Les en-têtes de colonnes utilisées ici (`TC_WEIGHT`, `TC_PRE_REQUISITE`,
`TC_OWNER_PATH`, etc.) — **si tu changes d'instance ou de
version Squash TM, ré-exporte un dossier existant et compare les en-têtes**
avant de faire confiance à ce générateur.

## ⚙️ Installation

```bash
git clone https://github.com/a-ringwater/squash-tc-gen.git
cd squash-tc-gen
go mod tidy
go build -o squash-tc-gen .
```

Prérequis : Go >= 1.22.

## 🚀 Usage

```bash
./squash-tc-gen -in examples/example.yaml -out import.xlsx
```

| Flag | Description | Défaut |
|---|---|---|
| `-in`  | Chemin du fichier YAML d'entrée (obligatoire) | — |
| `-out` | Chemin du fichier `.xlsx` généré | `import.xlsx` |

## 📄 Format d'entrée (YAML)

Voir [`examples/example.yaml`](examples/example.yaml) pour un exemple complet.

```yaml
project: "Mon Projet"
folder_path: "/Mon Projet/Ma Feature/01 - Mon dossier/EX-00x - ..."
reference: "EX-00x"
requirement_path: ""   # optionnel — chemin de l'exigence côté espace Exigences

test_cases:
  - title: "[BR-1] ..."
    weight: HIGH        # HIGH | MEDIUM | LOW
    prereq: "..."
    steps:
      - action: "..."
        expected: "..."
```

### Cas avancé : cas de test paramétrés

Pour regrouper plusieurs Business Rules à logique identique (ne différant
que par une valeur) dans un seul cas de test, plutôt que de dupliquer le
cas de test, ajoute `parameters` et `datasets` :

```yaml
    parameters: ["MON_PARAM"]
    datasets:
      - name: "Cas A"
        values: { MON_PARAM: "valeur A" }
      - name: "Cas B"
        values: { MON_PARAM: "valeur B" }
```

Utilise `{MON_PARAM}` dans `action`/`expected`/`prereq` pour référencer un
paramètre — Squash TM substituera la valeur au moment de l'exécution du
jeu de données correspondant.

## 📥 Import dans Squash TM

1. Génère le fichier avec cette CLI.
2. Dans Squash TM, espace **Cas de test** > bouton **Importer/Exporter** >
   **Importer**.
3. **Lance la simulation** avant de valider, pour repérer tout conflit de
   nommage dans le dossier cible.
4. Valide l'import.

## 🗂️ Structure du projet

```
.
├── main.go                     # CLI : flags, lecture YAML, validation, orchestration
├── internal/
│   ├── model/model.go          # Structures de données (Suite, TestCase, Step, Dataset)
│   └── generator/generator.go  # Génération du .xlsx via excelize
└── examples/
    └── example.yaml            # Exemple d'entrée
```

## ⚠️ Limites connues

- Une seule exigence liée par suite de test (pas de `LINK_REQ_TC`
  multi-exigences).
- Les champs `TC_NATURE`, `TC_TYPE`, `TC_MILESTONE`, `TC_AUTOMATABLE` sont laissés vides pour éviter une erreur d'import si les listes personnalisées diffèrent d'un projet à l'autre.

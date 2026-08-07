# squash-tc-gen

CLI Go qui génère un fichier Excel prêt à importer dans Squash TM (format
`TEST_CASES` / `STEPS` / `PARAMETERS` / `DATASETS` / `LINK_REQ_TC`) à partir
d'un fichier YAML décrivant les cas de test d'une story.

Projet réalisé pour automatiser la rédaction de cas de test manuels
(feature "SUT Registry") tout en apprenant Go.

## Pourquoi

Écrire les cas de test à la main dans l'IHM Squash TM est répétitif, et le
copier-coller depuis un tableau externe est source d'erreurs (chemins,
en-têtes de colonnes, HTML des champs riches). Ce générateur prend une
description structurée en YAML et produit directement le fichier d'import
au bon format.

Les en-têtes de colonnes utilisées ici (`TC_WEIGHT`, `TC_PRE_REQUISITE`,
`TC_OWNER_PATH`, etc.) ont été confirmées en exportant un dossier existant
depuis une instance Squash TM réelle — **si tu changes d'instance ou de
version Squash TM, ré-exporte un dossier existant et compare les en-têtes**
avant de faire confiance à ce générateur.

## Installation

```bash
git clone <ton-repo>
cd squash-tc-gen
go mod tidy
go build -o squash-tc-gen .
```

Prérequis : Go >= 1.22.

## Usage

```bash
./squash-tc-gen -in examples/sut-002.yaml -out import_SUT-002.xlsx
```

| Flag | Description | Défaut |
|---|---|---|
| `-in`  | Chemin du fichier YAML d'entrée (obligatoire) | — |
| `-out` | Chemin du fichier `.xlsx` généré | `import.xlsx` |

## Format d'entrée (YAML)

Voir [`examples/sut-002.yaml`](examples/sut-002.yaml) pour un exemple complet.

```yaml
project: "Gazelle v2"
folder_path: "/Gazelle v2/Sut registry/01 - My organization's SUT/SUT-00x - ..."
reference: "SUT-00x"
requirement_path: ""   # optionnel — chemin de l'exigence côté espace Exigences

test_cases:
  - title: "[BR-1] ..."
    weight: HIGH        # HIGH | MEDIUM | LOW
    prereq: "..."
    steps:
      - action: "..."
        expected: "..."

    # Optionnel : pour regrouper plusieurs Business Rules à logique
    # identique (ne différant que par une valeur) dans un seul cas de
    # test paramétré, plutôt que de dupliquer le cas de test.
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

## Import dans Squash TM

1. Génère le fichier avec cette CLI.
2. Dans Squash TM, espace **Cas de test** > bouton **Importer/Exporter** >
   **Importer** (réservé aux profils Administrateur / Chef de projet).
3. **Lance la simulation** avant de valider, pour repérer tout conflit de
   nommage dans le dossier cible.
4. Valide l'import.

## Note réseau (contournement, pas à reproduire en usage normal)

Ce projet a été développé dans un environnement sandbox n'autorisant pas
l'accès à `proxy.golang.org` ni `golang.org`. Le `go.mod` contient donc des
directives `replace` redirigeant certaines dépendances transitives
(`golang.org/x/*`, `gopkg.in/*`) vers leurs miroirs GitHub. **Sur une
machine avec un accès internet normal, ces `replace` ne sont probablement
pas nécessaires** — tu peux les supprimer et relancer `go mod tidy` pour
obtenir un `go.mod` "propre". Elles ne posent aucun problème si tu les
gardes, cela dit.

## Structure du projet

```
.
├── main.go                     # CLI : flags, lecture YAML, validation, orchestration
├── internal/
│   ├── model/model.go          # Structures de données (Suite, TestCase, Step, Dataset)
│   └── generator/generator.go  # Génération du .xlsx via excelize
└── examples/
    └── sut-002.yaml            # Exemple d'entrée
```

## Limites connues / pistes d'évolution

- `TC_NATURE`, `TC_TYPE`, `TC_MILESTONE`, `TC_AUTOMATABLE` ne sont pas
  renseignés (laissés vides) pour éviter une erreur d'import si les codes
  de listes personnalisées diffèrent d'un projet à l'autre. À ajouter dans
  le modèle si besoin.
- Pas de génération automatique de `LINK_REQ_TC` multi-exigences (une seule
  exigence par suite pour l'instant).
- Pas de tests unitaires Go pour l'instant (`generator_test.go` à écrire).

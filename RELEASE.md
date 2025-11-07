# Guide de Release

Ce document explique comment créer une nouvelle release de `cloud_helper`.

## Prérequis

- Accès en écriture au repository GitHub
- Git configuré avec authentification SSH ou token GitHub
- Toutes les modifications commitées et pushées sur `main`

## Processus de Release

### 1. Vérifier l'état du repository

```bash
# S'assurer d'être sur main et à jour
git checkout main
git pull origin main

# Vérifier qu'il n'y a pas de modifications non commitées
git status
```

### 2. Créer et pousser un tag de version

Les releases sont automatiquement déclenchées par la création d'un tag commençant par `v`.

```bash
# Créer un tag annoté (format: vX.Y.Z)
git tag -a v1.0.0 -m "Release v1.0.0

Description des changements principaux:
- Fonctionnalité 1
- Fonctionnalité 2
- Fix important"

# Pousser le tag vers GitHub
git push origin v1.0.0
```

### 3. GitHub Actions prend le relais

Une fois le tag poussé, le workflow `.github/workflows/release.yml` s'exécute automatiquement :

1. **Build** : Compile les binaires pour Linux, Windows et macOS
2. **Packages** : Crée des packages `.deb`, `.rpm` et `.apk`
3. **Release** : Publie une release GitHub avec tous les artifacts
4. **Notes** : Génère automatiquement les notes de release depuis les commits

### 4. Vérifier la release

1. Aller sur https://github.com/robinportigliatti/cloud_helper/releases
2. Vérifier que la release est créée
3. Tester le téléchargement d'un binaire

## Gestion sémantique des versions

Suivre le [Semantic Versioning](https://semver.org/lang/fr/) :

- **v1.0.0** → Release majeure (breaking changes)
- **v0.1.0** → Nouvelles fonctionnalités (backward compatible)
- **v0.0.1** → Corrections de bugs

## En cas de problème

### La release a échoué

1. Consulter les logs dans l'onglet "Actions" de GitHub
2. Corriger le problème
3. Supprimer le tag local et distant :
   ```bash
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```
4. Recréer le tag après correction

### Modifier une release existante

Les releases GitHub peuvent être éditées manuellement depuis l'interface web pour :
- Modifier les notes de release
- Ajouter/supprimer des artifacts
- Marquer comme pre-release

## Configuration goreleaser

Le fichier `.goreleaser.yaml` contrôle le processus de build et release :
- Plateformes cibles
- Formats de packages
- Métadonnées (mainteneur, description, etc.)

Consulter la [documentation goreleaser](https://goreleaser.com/customization/) pour plus d'options.

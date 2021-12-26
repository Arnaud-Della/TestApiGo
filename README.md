# TestApiGo

Premiere API en Go avec Gorilla (https://github.com/gorilla/mux) et MangoDB 3.0.14
L' api est accessible à l'adresse http://nodetest.ddns.net:8080/

On peut afficher toutes les taches :
  GET /Tasks
On peut ajouter une nouvelle tache :
  POST /Task
( Si l'ajout ne comprend pas tous les champs requis => Erreur 400 )
On peut modifier une tache :
  PUT /Task/{id}
On peut supprimer une tache :
  DELETE /Task/{id}
On peut afficher une tache en particulier :
  GET /Task/{id}
On peut affciher l'aide :
  GET /Help
On peut filtrer les tache avec des criteres de recherches :
  GET /Tasks/search
  

package db

import (
	"context"
	_ "os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/jak103/usu-gdsf/config"
	"github.com/jak103/usu-gdsf/log"
	"github.com/jak103/usu-gdsf/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ Database = (*Firestore)(nil)

type Firestore struct {
	client *firestore.Client
}

// RemoveGame removes the given game from the db
func (db Firestore) RemoveGame(game models.Game) error {
	// query

	snapShot, err := db.client.Collection("games").Doc(game.Id).Get(context.Background())
	if err != nil {
		log.WithError(err).Error("Firestore query error in RemoveGame")
	}

	// delete doc
	_, err = snapShot.Ref.Delete(context.Background())
	if err != nil {
		log.WithError(err).Error("Firestore deletion error in RemoveGame")
		return err
	}

	return nil
}

func (db Firestore) RemoveGameByTag(tag string) error {
	return nil
}

func (db Firestore) SortGames(field_name string, order int) ([]models.Game, error) {
	return nil, nil
}

// GetGamesByTag search and return all games with given tag
func (db Firestore) GetGamesByTags(tags []string, matchAll bool) ([]models.Game, error) {
	// query
	gc := db.client.Collection("games")
	operator := ""
	if matchAll {
		operator = "array-contains"
	} else {
		operator = "array-contains-any"
	}

	result := gc.Where("Tags", operator, tags)

	// get docs
	docs, err := result.Documents(context.Background()).GetAll()
	if err != nil {
		log.WithError(err).Error("Firestore query error in GetGamesByTags")
		return []models.Game{}, err
	}

	// decode
	games := make([]models.Game, len(docs))
	for i, doc := range docs {
		game := models.Game{}
		err = doc.DataTo(&game)
		game.Id = doc.Ref.ID
		if err != nil {
			log.WithError(err).Error("Firestore decode error in GetGamesByTag")
			return []models.Game{}, err
		}
		games[i] = game
	}

	return games, nil
}

// GetGameByFirstLetter find and return the game with the given First Letter
func (db Firestore) GetGamesByFirstLetter(letter string) ([]models.Game, error) {
	games := make([]models.Game, 0)
	gc := db.client.Collection("games")

	documents := gc.DocumentRefs(context.Background())
	for {
		docRef, docRefErr := documents.Next()
		if docRefErr == iterator.Done {
			break
		}

		var game models.Game

		if docSnapshot, _ := docRef.Get(context.Background()); docSnapshot != nil {
			_ = docSnapshot.DataTo(&game)
			game.Id = docRef.ID
		}
		if strings.EqualFold(game.Name[0:1], letter) {
			games = append(games, game)
		}
	}
	return games, nil
}

// GetGameByID find and return the game with the given db ID
func (db Firestore) GetGameByID(id string) (models.Game, error) {
	snapShot, err := db.client.Collection("games").Doc(id).Get(context.Background())
	if status.Code(err) == codes.NotFound {
		return models.Game{}, err
	}
	game := models.Game{}
	convErr := snapShot.DataTo(&game)
	game.Id = snapShot.Ref.ID
	if convErr != nil {
		log.WithError(convErr).Error("Cannot convert firestore snapshot to game struct")
	}
	return game, nil
}

func (db Firestore) GetDownloadByID(id string) (models.Download, error) {
	snapShot, err := db.client.Collection("downloads").Doc(id).Get(context.Background())
	if status.Code(err) == codes.NotFound {
		return models.Download{}, err
	}
	download := models.Download{}
	convErr := snapShot.DataTo(&download)
	download.Id = snapShot.Ref.ID
	if convErr != nil {
		log.WithError(convErr).Error("Cannot convert firestore snapshot to download struct")
	}
	return download, nil

}

// AddGame Add a new game to the remote database. Returns unique game ID
func (db Firestore) AddGame(game models.Game) (string, error) {
	docRef, _, err := db.client.Collection("games").Add(context.Background(), game)

	if err != nil {
		log.WithError(err).Error("Failed to add game to firestore db")
		return docRef.ID, err
	}
	return docRef.ID, nil
}

func (db Firestore) AddDownload(download models.Download) (string, error) {
	docRef, _, err := db.client.Collection("downlaods").Add(context.Background(), download)

	if err != nil {
		log.WithError(err).Error("Failed to add download to firestore db")
		return docRef.ID, err
	}
	return docRef.ID, nil
}

func (db Firestore) GetAllGames() ([]models.Game, error) {
	games := make([]models.Game, 0)
	gc := db.client.Collection("games")

	documents := gc.DocumentRefs(context.Background())
	for {
		docRef, docRefErr := documents.Next()

		if docRefErr == iterator.Done {
			break
		}

		var game models.Game

		if docSnapshot, _ := docRef.Get(context.Background()); docSnapshot != nil {
			_ = docSnapshot.DataTo(&game)
			game.Id = docRef.ID
		}

		games = append(games, game)
	}

	return games, nil
}

func (db Firestore) GetAllDownloads() ([]models.Download, error) {
	downloads := make([]models.Download, 0)
	gc := db.client.Collection("downloads")

	documents := gc.DocumentRefs(context.Background())
	for {
		docRef, docRefErr := documents.Next()

		if docRefErr == iterator.Done {
			break
		}

		var download models.Download

		if docSnapshot, _ := docRef.Get(context.Background()); docSnapshot != nil {
			_ = docSnapshot.DataTo(&download)
			download.Id = docRef.ID
		}

		downloads = append(downloads, download)
	}

	return downloads, nil
}

func (db Firestore) UpdateGame(newGameInfo models.Game, id string) (models.Game, error) {
	_, err := db.client.Collection("games").Doc(id).Set(context.Background(), map[string]interface{}{
		"Name":         newGameInfo.Name,
		"Developer":    newGameInfo.Developer,
		"Version":      newGameInfo.Version,
		"DownloadLink": newGameInfo.DownloadLink,
	}, firestore.MergeAll)
	if err != nil {
		log.WithError(err).Error("Failed to update game in firestore db")
		return newGameInfo, err
	}
	return newGameInfo, nil
}

func (db Firestore) CreateUser(newUser models.User) (string, error) {
	userDocRef, _, err := db.client.Collection("users").Add(context.Background(), newUser)
	if err != nil {
		log.WithError(err).Error("Failed to insert new user into Firestore DB")
		return userDocRef.ID, err
	}
	return userDocRef.ID, nil
}

// Disconnect disconnects from the remote database
func (db *Firestore) Disconnect() error {
	// Close the client connection if it is open
	if db.client != nil {
		if err := db.client.Close(); err != nil {
			log.WithError(err).Error("Failed to disconnect firestore")
			return err
		}
	}

	return nil
}

// Connect allows the user to connect to the database
func (db *Firestore) Connect() error {
	// Sets your Google Cloud Platform project ID.
	projectID := config.FirestoreProjectId

	client, err := firestore.NewClient(context.Background(), projectID)
	if err != nil {
		log.WithError(err).Error("Failed to get firestore client")
		return err
	}

	// Etablish Database Collection object
	db.client = client

	return nil
}

// AddReview Add a new review to the remote database and update average score. Returns unique review ID
func (db Firestore) AddReview(review models.Review) (string, error) {
	docRef, _, err := db.client.Collection("reviews").Add(context.Background(), review)

	if err != nil {
		log.WithError(err).Error("Failed to add review to firestore db")
		return docRef.ID, err
	}
	game, err := db.GetGameByID(review.GameId)
	game.ReviewIds = append(game.ReviewIds, review.Id)

	totalScore := game.AverageReview * (float64((len(game.ReviewIds) - 1)))
	game.AverageReview = ((totalScore + review.Score) / (float64(len(game.ReviewIds))))

	return docRef.ID, nil
}

// GetReviewByID find and return the Review with the given db ID
func (db Firestore) GetReviewByID(id string) (models.Review, error) {
	snapShot, err := db.client.Collection("reviews").Doc(id).Get(context.Background())
	if status.Code(err) == codes.NotFound {
		return models.Review{}, err
	}
	review := models.Review{}
	convErr := snapShot.DataTo(&review)
	review.Id = snapShot.Ref.ID
	if convErr != nil {
		log.WithError(convErr).Error("Cannot convert firestore snapshot to review struct")
	}
	return review, nil
}

// RemoveReview removes the given review from the db
func (db Firestore) RemoveReview(review models.Review) error {
	// query

	snapShot, err := db.client.Collection("reviews").Doc(review.Id).Get(context.Background())
	if err != nil {
		log.WithError(err).Error("Firestore query error in RemoveReview")
	}

	// delete doc
	_, err = snapShot.Ref.Delete(context.Background())
	if err != nil {
		log.WithError(err).Error("Firestore deletion error in RemoveReview")
		return err
	}

	return nil
}

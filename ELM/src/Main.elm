module Main exposing (..)


import Browser
import Http exposing (..)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onInput, onClick)
import Random
import List exposing (drop)
import Decoder exposing (..)



-- MAIN


main =
  Browser.element
    { init = init
    , update = update
    , subscriptions = subscriptions
    , view = view
    }



-- MODEL


type Model
  = Failure
  | Loading
  | Success String
  | FullText String
  | Word String
  | Def (List Definition)


init : () -> (Model, Cmd Msg)
init _ =
  ( Loading
  , Http.get
      { url = "https://raw.githubusercontent.com/ninarodriguezz/ELP/main/ELM/thousand_words_things_explainer.txt"
      , expect = Http.expectString GotText
      }
  )



-- UPDATE


type Msg
  = GotText (Result Http.Error String)
  | RandomInt Int
  | GotJson (Result Http.Error (List Definition))
  | CheckWord String
  | ShowWord
  | HideWord



getWordAtIndex : Int -> List String -> String
getWordAtIndex index input =
    case drop (index-1) input of
      (x::xs) -> x
      [] -> "err"


getJson : String -> Cmd Msg
getJson word = 
  Http.get
    { url = "https://api.dictionaryapi.dev/api/v2/entries/en/" ++ word
    , expect = Http.expectJson GotJson decoderJson
    }      


update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    GotText result ->
      case result of
        Ok fullText ->
          (FullText fullText, Random.generate RandomInt (Random.int 0 (List.length (String.words fullText))))  
        Err _ ->
          (Failure, Cmd.none)

    GotJson result ->
      case result of
        Ok def ->
          (Def def, Cmd.none)
        Err _ ->  
          (Failure, Cmd.none)
         

    RandomInt number -> case model of
      FullText text -> (Word (getWordAtIndex number (String.words text)), Random.generate RandomInt (Random.int 0 10))  
      Word word -> ( Loading
        , getJson word
        )
      Failure -> (model, Cmd.none)
      Loading -> (model, Cmd.none)
      Success text -> (model, Cmd.none)
      Def text -> (model, Cmd.none)

    CheckWord word -> 
      case model of
        Failure -> (model, Cmd.none)
        Loading -> (model, Cmd.none)
        Success text -> (model, Cmd.none)
        FullText text -> (model, Cmd.none)
        Word text -> (model, Cmd.none)
        Def def -> case def of
          (x::_) -> if x.word == word then 
            ( Def (Definition "Guessed" [] :: def), Cmd.none)
            else 
            ( Def (Definition "Failed" [] :: def), Cmd.none)
          [] -> (model, Cmd.none)  

    ShowWord -> (model, Cmd.none)
    HideWord -> (model, Cmd.none)      
    


-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
  Sub.none



-- VIEW


view : Model -> Html Msg
view model =
  let 
    viewDef def =
      li [] [ text def ]

    viewMeaning meaning = ul []
      [ li [] 
        [ text meaning.partOfSpeech
        , pre [] []
        , ol [] (List.map viewDef meaning.definition) 
        , pre [] [ text " " ]
        ]
      ]

    viewGame def = div []
      [ h1 [] [ text "Guess the Word Game" ]
      , h3 [] [ text "Meanings :" ]
      , ul [] (List.map viewMeaning def.definition)
      , input [ placeholder "Enter a word",  onInput CheckWord ] []
      , div [] 
        [ pre [] []
        , button [ onClick ShowWord ] [ text "Show Answer" ]
        , button [ onClick HideWord ] [ text "Hide Answer" ]
        ]
      ]

    viewGuessed def = div []
      [ h1 [] [ text "Guess the Word Game" ]
      , h3 [] [ text "Meanings :" ]
      , ul [] (List.map viewMeaning def.definition)
      , text ("Exactly ! The word to find was " ++ def.word ++ " !")
      ]

    viewFailed def = div []
      [ h1 [] [ text "Guess the Word Game" ]
      , h3 [] [ text "Meanings :" ]
      , ul [] (List.map viewMeaning def.definition)
      , text "You failed to guess the word, you can try again."
      , input [ placeholder "Enter a word", onInput CheckWord ] []
      ]  

  in
  case model of
    Failure ->
      text "I was unable to load your book."

    Loading ->
      text "Loading..."

    Success texte ->
      pre [] [ text texte ]

    Word word ->
      pre [] [ text word ]

    FullText texte ->
      pre [] [ text texte ]  

    Def def -> case def of
      (x::_) -> viewGame x
      [] -> pre[] [ text "error" ]

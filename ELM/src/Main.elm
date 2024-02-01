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
  | FullText String
  | Word String
  | Def (List Definition)
  | View { content : String, def : (List Definition) }


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
      Def text -> (model, Cmd.none)
      View map -> (model, Cmd.none)

    CheckWord word -> 
      case model of
        Failure -> (model, Cmd.none)
        Loading -> (model, Cmd.none)
        FullText text -> (model, Cmd.none)
        Word text -> (model, Cmd.none)
        View map -> case map.def of
          (x::_) -> if x.word == word then 
            ( View { content = "Guessed", def = map.def}, Cmd.none)
            else 
            (model, Cmd.none)
          [] -> (model, Cmd.none)   
        Def def -> case def of
          (x::_) -> if x.word == word then 
            ( View { content = "Guessed", def = def}, Cmd.none)
            else 
            (model, Cmd.none)
          [] -> (model, Cmd.none)  

    ShowWord -> 
      case model of
        Failure -> (model, Cmd.none)
        Loading -> (model, Cmd.none)
        FullText text -> (model, Cmd.none)
        Word text -> (model, Cmd.none)
        View map -> ( View { content = "Show", def = map.def}, Cmd.none)
        Def def -> ( View { content = "Show", def = def}, Cmd.none) 

    HideWord ->   
      case model of
        Failure -> (model, Cmd.none)
        Loading -> (model, Cmd.none)
        FullText text -> (model, Cmd.none)
        Word text -> (model, Cmd.none)
        View map -> ( View { content = "Hide", def = map.def}, Cmd.none)
        Def def -> ( View { content = "Hide", def = def}, Cmd.none)


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

    viewGame def = div [ style "margin" "20px" ]
      [ h1 [] [ text "Guess the Word !" ]
      , h3 [] [ text "Meanings :" ]
      , ul [] (List.map viewMeaning def.definition)
      , input [ placeholder "Enter a word", onInput CheckWord ] []
      , div [] 
        [ pre [] []
        , button [ onClick ShowWord ] [ text "Show Answer" ]
        , button [ onClick HideWord ] [ text "Hide Answer" ]
        ]
      ]

    viewGuessed def = div [ style "margin" "20px" ]
      [ h1 [] [ text "Guess the Word !" ]
      , h3 [] [ text "Meanings :" ]
      , ul [] (List.map viewMeaning def.definition)
      , text ("Exactly ! The word to guess was " ++ def.word ++ " !")
      ]

    viewWord def = div [  style "margin" "20px" ]
      [ h1 [] [ text "Guess the Word !" ]
      , h2 [] [ text ("The word to guess is : " ++ def.word) ]
      , h3 [] [ text "Meanings :" ]
      , ul [] (List.map viewMeaning def.definition)
      , div [] 
        [ pre [] []
        , button [ onClick ShowWord ] [ text "Show Answer" ]
        , button [ onClick HideWord ] [ text "Hide Answer" ]
        ]
      ]

 
  in
  case model of
    Failure ->
      text "I was unable to load your book."

    Loading ->
      text "Loading..."

    Word word ->
      pre [] [ text word ]

    FullText texte ->
      pre [] [ text texte ]  

    Def def -> case def of
      (x::_) -> viewGame x
      [] -> pre[] [ text "error1" ]

    View map -> case map.content of

      "Guessed" -> case map.def of
        (x::_) -> viewGuessed x
        [] -> pre[] [ text "error2" ]

      "Show" -> case map.def of
        (x::_) -> viewWord x
        [] -> pre[] [ text "error3" ]

      "Hide" -> case map.def of
        (x::_) -> viewGame x
        [] -> pre[] [ text "error4" ] 

      _ -> pre[] [ text "error5" ]  
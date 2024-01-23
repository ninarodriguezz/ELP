-- Make a GET request to load a book called "Public Opinion"
--
-- Read how it works:
--   https://guide.elm-lang.org/effects/http.html
--

import Browser
import Html exposing (Html, text, pre, div)
import Http
import Random
import List exposing (drop)



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


getWordAtIndex : Int -> List String -> String
getWordAtIndex index input =
    case drop (index-1) input of
      (x::xs) -> x
      [] -> "err"


update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    GotText result ->
      case result of
        Ok fullText ->
          (Success fullText, Random.generate RandomInt (Random.int 0 (List.length (String.words fullText))))  

        Err _ ->
          (Failure, Cmd.none)

    RandomInt number -> case model of
      Success text -> (Success (getWordAtIndex number (String.words text)), Cmd.none)   
      Failure -> (model, Cmd.none)
      Loading -> (model, Cmd.none)



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
  Sub.none



-- VIEW


view : Model -> Html Msg
view model =
  case model of
    Failure ->
      text "I was unable to load your book."

    Loading ->
      text "Loading..."

    Success fullText ->
      pre [] [ text fullText ]
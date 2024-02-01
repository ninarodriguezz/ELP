-- HTTP

type alias Meaning =
    {partOfSpeech : String,
    definition : (List String)}

type alias Definition =
    { word : String
    , definition : List Meaning}    

getJson : String -> Cmd Msg
getJson word = 
  Http.get
    { url = "https://api.dictionaryapi.dev/api/v2/entries/en/" ++ word
    , expect = Http.expectJson GotJson decoderJson
    }

decoderJson : Decoder (List Definition)
decoderJson =
  list decoderDef

decoderDef : Decoder Definition
decoderDef =
  map2 Definition
    (field "word" string)
    (field "meanings" (list decoderMeaning))

decoderMeaning : Decoder Meaning
decoderMeaning =
  map2 Meaning
    (field "partOfSpeech" string)
    (field "definitions" (list decoderString))

decoderString : Decoder String 
decoderString =
    field "definition" string


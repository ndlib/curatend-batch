# Register and configure remote identifiers for persisted objects
Hydra::RemoteIdentifier.configure do |config|
  doi_credentials = Psych.load_file(Rails.root.join("config/doi.yml").to_s).fetch(Rails.env)
  config.remote_service(:doi, doi_credentials) do |doi|
    doi.register(SeniorThesis) do |map|
      map.target {|obj| Curate.permanent_url_for(obj) }
      map.creator :creator
      map.title :title
      map.publisher {|o| Array(o.publisher).join("; ")}
      map.publicationyear {|o| o.date_uploaded.year }
      map.set_identifier {|o,value| o.identifier = value; o.save }
    end

    doi.register(Dataset, Image, Document, Article, Etd) do |map|
      map.target {|obj| Curate.permanent_url_for(obj) }
      map.creator {|obj| Array.wrap(obj.creator).collect(&:to_s).join(", ") }
      map.title :title
      map.publisher {|o| Array.wrap(o.publisher).join("; ")}
      map.publicationyear {|o| o.date_uploaded.year }
      map.set_identifier {|o,value| o.identifier = value; o.save }
    end

  end
end

Curate.configure do |config|
  # Injected via `rails g curate:work SeniorThesis`
  config.register_curation_concern :senior_thesis
  config.register_curation_concern :finding_aid
  config.register_curation_concern :dataset
  config.register_curation_concern :article
  config.register_curation_concern :image
  config.register_curation_concern :document
  config.register_curation_concern :etd

  config.application_root_url = Rails.configuration.application_root_url
end

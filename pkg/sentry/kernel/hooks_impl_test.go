package kernel

func testBuildContexts() ScriptContexts {
	builder := ScriptContextsBuilderOf()
	builder = builder.AddContext3(HooksJsName, &IndependentHookAddableAdapter{ht: jsRuntime.hooksTable})
	builder = builder.AddContext3(JsPersistenceContextName,
		&ObjectAddableAdapter{name: JsGlobalPersistenceObject, object: jsRuntime.Global})
	return builder.Build()
}

package runners;

import io.cucumber.testng.AbstractTestNGCucumberTests;
import io.cucumber.testng.CucumberOptions;

@CucumberOptions(
    features = "src/test/resources/features",
    glue = "steps",
    plugin = {"pretty", "html:target/cucumber-report.html"},
    monochrome = true,
    dryRun = true,
   tags="@smoke"
)
public class TestRunner extends AbstractTestNGCucumberTests {
}

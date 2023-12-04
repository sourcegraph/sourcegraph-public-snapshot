// Original:
// repo:^github\.com/radzenhq/radzen-blazor$ file:^Radzen\.Blazor\.Tests/DataGridTests\.cs

using AngleSharp.Dom;
using Bunit;
using Microsoft.AspNetCore.Components;
using Microsoft.AspNetCore.Components.Rendering;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text.RegularExpressions;
using Xunit;
namespace Radzen.Blazor.
{
    public class DataGridTests
    {
        // Css classes tests
        [Fact]
        public void DataGrid_Renders_CssClass()
        {
            using var ctx = new TestContext();
            ctx.JSInterop.Mode = JSRuntimeMode.Loose;
            ctx.JSInterop.SetupModule("_content/Radzen.Blazor/Radzen.Blazor.js");
            var component = ctx.RenderComponent<RadzenGrid<dynamic>>(parameterBuilder =>
            {
                parameterBuilder.Add<IEnumerable<dynamic>>(p => p.Data, new[] { new { Id = 1 }, new { Id = 2 }, new { Id = 3 } });
                parameterBuilder.Add<RenderFragment>(p => p.Columns, builder =>
                {
                    builder.OpenComponent(0, typeof(RadzenGridColumn<dynamic>));
                    builder.AddAttribute(1, "Property", "Id");
                    builder.CloseComponent();
                });
            });

            // Main
            Assert.Contains(@$"rz-datatable-scrollable-wrapper", component.Markup);
            Assert.Contains(@$"rz-datatable-scrollable-view", component.Markup);

            var component = ctx.RenderComponent<RadzenGrid<dynamic>>(parameterBuilder =>
            {
                parameterBuilder.Add<IEnumerable<dynamic>>(p => p.Data, new[] { new { Id = 1 }, new { Id = 2 }, new { Id = 3 } });
                parameterBuilder.Add<RenderFragment>(p => p.Columns, builder =>
                {
                    builder.OpenComponent(0, typeof(RadzenGridColumn<dynamic>));
                    builder.AddAttribute(1, "Property", "Id");
                    builder.CloseComponent();
                });
            });

            var markup = new Regex(@"\s\s+").Replace(component.Markup, "").Trim();
            Assert.Contains(@$"""rz-cell-data"">1", markup);
            Assert.Contains(@$"""rz-cell-data"">2", markup);
            Assert.Contains(@$"""rz-cell-data"">3", markup);
            Assert.Contains(@"
Lorem Ipsum
", component.Markup);
        }
    }
}
